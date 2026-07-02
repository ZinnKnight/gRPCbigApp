package App

import (
	"context"
	"encoding/json"
	"fmt"
	clientPG "gRPCbigapp/App/ClientService/Adapter/PostgresAdapter"
	clientGRPC "gRPCbigapp/App/ClientService/Adapter/grpcAdapter"
	clientUC "gRPCbigapp/App/ClientService/UseCase"
	orderPG "gRPCbigapp/App/OrderService/Adapters/Postgres"
	orderGRPC "gRPCbigapp/App/OrderService/Adapters/grpcAdapter"
	"gRPCbigapp/App/OrderService/Streaming"
	orderUC "gRPCbigapp/App/OrderService/UseCase"
	"gRPCbigapp/Proto/protoPB"
	authAdapter "gRPCbigapp/Shared/Auth/AuthAdapter"
	authInterceptor "gRPCbigapp/Shared/Auth/AuthInterceptor"
	"gRPCbigapp/Shared/Config"
	"gRPCbigapp/Shared/ErrorInterceptor"
	"gRPCbigapp/Shared/Events"
	"gRPCbigapp/Shared/Idempotentor"
	Kafka "gRPCbigapp/Shared/Kafka"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	Metrics2 "gRPCbigapp/Shared/Metrics"
	"gRPCbigapp/Shared/Orchestrator"
	"gRPCbigapp/Shared/Outbox"
	"gRPCbigapp/Shared/PanicInterceptor"
	"gRPCbigapp/Shared/Policy"
	"gRPCbigapp/Shared/Quota"
	RateLimiter2 "gRPCbigapp/Shared/RateLimiter"
	redisClient "gRPCbigapp/Shared/Redis"
	"gRPCbigapp/Shared/SagaMessages"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/Shared/Txmanager"
	"gRPCbigapp/Shared/ValidationIntercepter"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	pgPool "gRPCbigapp/Shared/Postgres"
	otlpexporter "gRPCbigapp/Shared/Tracing/OPTLExp"
)

// hardcap shutdown
const shutdownTimeout = 15 * time.Second

// hardcap ttl for jwt
const jwtTTL = 4 * time.Hour

type App struct {
	cfg             *Config.Config
	logger          LoggerPorts.Logger
	pool            *pgxpool.Pool
	rdb             *redisClient.RedisClient
	producer        *Kafka.Producer
	relay           *Outbox.Relay
	orderConsumer   *Kafka.Consumer
	orchestrator    *Orchestrator.Orchestrator
	statusConsumer  *Kafka.Consumer
	hub             *Streaming.Hub
	grpcServer      *grpc.Server
	listening       net.Listener
	metricsHandler  http.Handler
	tracingShutdown tracing.ShutDownTracing
}

func New(ctx context.Context, cfg *Config.Config, logger LoggerPorts.Logger) (*App, error) {
	app := &App{
		cfg:    cfg,
		logger: logger,
	}

	var traceExp sdktrace.SpanExporter

	if cfg.TracingEnabled {
		exp, err := otlpexporter.NewgRPCExporter(ctx, cfg.OpenTelemetryEndpoint)
		if err != nil {
			return nil, fmt.Errorf("orderapp, otlpexporter: %w", err)
		}
		traceExp = exp
	}

	tracingShutDown, err := tracing.Init(ctx, tracing.Config{
		Logger:         logger,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		SampleRatio:    cfg.TracingSampleRatio,
	}, traceExp)
	if err != nil {
		return nil, fmt.Errorf("orderapp, tracing initiation: %w", err)
	}
	app.tracingShutdown = tracingShutDown

	pool, err := pgPool.NewPool(ctx, pgPool.Config{
		DatabaseURL:    cfg.DatabaseURL,
		MaxConnections: int32(cfg.DBMaxConn),
		MinConnections: int32(cfg.DBMinConn),
		MaxConnTTL:     time.Duration(cfg.DBMaxConnTTL) * time.Minute,
		DBMaxConnIdTTL: time.Duration(cfg.DBMaxConnIdTTL) * time.Minute,
		AfterConn: func(ctx context.Context, conn *pgx.Conn) error {
			pgxdecimal.Register(conn.TypeMap())
			return nil
		},
	})
	if err != nil {
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, postgres pool initiation: %w", err)
	}
	app.pool = pool

	rdb, err := redisClient.NewRedisClient(ctx, redisClient.Config{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	if err != nil {
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, redis pool initiation: %w", err)
	}
	app.rdb = rdb

	producer, err := Kafka.NewProducer(ctx, Kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		ClientID: cfg.ServiceName,
	})
	if err != nil {
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, Kafka producer initiation: %w", err)
	}
	app.producer = producer

	txManager := Txmanager.NewTxManager(pool)

	topicResolver := func(e Events.Events) string {
		switch e.EventType {
		case SagaMessages.EventOrderCreated:
			return SagaMessages.TopicOrderEvents
		case SagaMessages.CommandReserveStock:
			return SagaMessages.TopicSagaCommands
		case SagaMessages.EventOrderStatusChanged:
			return SagaMessages.TopicOrderStatus
		default:
			return e.AggregationType + ".events"
		}
	}
	eventEmit := Outbox.NewWriter(pool, topicResolver)

	app.relay = Outbox.NewRelay(pool, logger, producer, 100, time.Second)

	orderRepo := orderPG.NewOrderRepo(pool)

	orchestrator := Orchestrator.NewOrchestrator(
		orderRepo,
		txManager,
		eventEmit,
		Idempotentor.NewGuard(pool, "order-orchestrator"),
		logger,
	)

	orderConsumer, err := Kafka.NewConsumer(ctx, Kafka.ConsumerConfig{
		Brokers: cfg.KafkaBrokers,
		Group:   "order-orchestrator",
		Topics:  []string{SagaMessages.TopicOrderEvents, SagaMessages.TopicSagaReplies},
	}, logger)
	if err != nil {
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, order-orchestrator initiation: %w", err)
	}
	app.orderConsumer = orderConsumer
	app.orchestrator = orchestrator

	app.hub = Streaming.NewHub()
	statusConsumer, err := Kafka.NewConsumer(ctx, Kafka.ConsumerConfig{
		Brokers:    cfg.KafkaBrokers,
		Group:      "order-status-sream" + uuid.NewString(),
		Topics:     []string{SagaMessages.TopicOrderStatus},
		StartAtEnd: true,
	}, logger)
	if err != nil {
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, order-status-sream initiation: %w", err)
	}
	app.statusConsumer = statusConsumer

	limiter := RateLimiter2.NewRateLimiter(rdb.Client, cfg.RateLimitPerMin, time.Minute)

	policyProvider, err := Policy.NewStaticProvider()
	if err != nil {
		app.statusConsumer.Close()
		app.orderConsumer.Close()
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, policy provider initiation: %w", err)
	}
	quotaEnforced := Quota.NewEnforced(policyProvider, limiter)

	orderUseCase := orderUC.NewOSUseCase(orderRepo, eventEmit, txManager, quotaEnforced, logger)

	clientRepo := clientPG.NewUserRepo(pool)
	clientUseCase := clientUC.NewUserUseCase(clientRepo, eventEmit, txManager, quotaEnforced, logger)

	metricsRecord := Metrics2.NewPrometheusRecord()
	app.metricsHandler = metricsRecord.Registry()
	jwtService := authAdapter.NewJWTService([]byte(cfg.JWTSecretKey), jwtTTL)

	orderHandler := orderGRPC.NewOrderHandler(logger, orderUseCase, app.hub)
	clientHandler := clientGRPC.NewUserhandler(clientUseCase, logger, jwtService, clientGRPC.NewPlanChangePreRequestStub())

	app.grpcServer = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			PanicInterceptor.PanicRecoveryInterceptor(logger),
			Metrics2.UnaryServerInterceptor(metricsRecord),
			ErrorInterceptor.UnaryServerInterceptor(logger),
			ValidationIntercepter.UnaryServerInterceptor(),
			authInterceptor.AuthInterceptor([]byte(cfg.JWTSecretKey)),
			RateLimiter2.UnaryServerInterceptor(limiter),
		),
	)
	protoPB.RegisterOrderServiceServer(app.grpcServer, orderHandler)
	protoPB.RegisterAuthServiceServer(app.grpcServer, clientHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		app.statusConsumer.Close()
		app.orderConsumer.Close()
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutdown(ctx)
		return nil, fmt.Errorf("orderapp, listener: %w", err)
	}
	app.listening = lis

	return app, nil
}

func (app *App) shutDown() {
	stoped := make(chan struct{})
	go func() {
		app.grpcServer.Stop()
		close(stoped)
	}()

	select {
	case <-stoped:
		app.logger.LogInfo("orderapp, grpc gracefull stop")
	case <-time.After(shutdownTimeout):
		app.logger.LogError("orderapp, grpc gracefull stop timeout, forced shutdown")
		app.grpcServer.Stop()
		<-stoped
	}

	if app.statusConsumer != nil {
		app.statusConsumer.Close()
	}
	if app.orderConsumer != nil {
		app.orderConsumer.Close()
	}
	if app.producer != nil {
		app.producer.Close()
	}

	app.pool.Close()
	if err := app.rdb.Close(); err != nil {
		app.logger.LogError("orderapp, redis close", LoggerPorts.Field{Key: "error", Value: err.Error()})
	}

	shCTX, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.tracingShutdown(shCTX); err != nil {
		app.logger.LogError("orderapp, tracing shutdown", LoggerPorts.Field{Key: "error", Value: err.Error()})
	}
}

func (app *App) publishStatusUpdate(_ context.Context, message Kafka.Message) error {
	if message.Header["event_type"] != SagaMessages.EventOrderStatusChanged {
		return nil
	}

	var p SagaMessages.OrderStatusChangedPayload

	if err := json.Unmarshal(message.Value, &p); err != nil {
		app.logger.LogError("orderApp, bad OrderStatusChange payload",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil
	}
	app.hub.Publish(Streaming.Update{OrderID: p.OrderID, Status: p.Status})
	return nil
}

func (app *App) Run(ctx context.Context) error {
	workerCTX, workerCancel := context.WithCancel(ctx)

	go func() {
		if err := Metrics2.StartMetricsServer(workerCTX, app.cfg.MetricsPort, app.metricsHandler); err != nil {
			app.logger.LogError("orderApp, metrics server startup", LoggerPorts.Field{Key: "error", Value: err})
		}
	}()

	// outbox
	go app.relay.Run(workerCTX)

	// оркестратор
	go app.orderConsumer.Run(workerCTX, app.orchestrator.Handle)

	// стриминг
	go app.statusConsumer.Run(workerCTX, app.publishStatusUpdate)

	serverErr := make(chan error, 1)

	go func() {
		app.logger.LogInfo("orderApp, server startup",
			LoggerPorts.Field{Key: "grpc_port", Value: app.cfg.GRPCPort},
			LoggerPorts.Field{Key: "metrics_port", Value: app.cfg.MetricsPort})
		if err := app.grpcServer.Serve(app.listening); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	var runErr error

	select {
	case <-ctx.Done():
		app.logger.LogInfo("orderApp, server shutdown")
	case err := <-serverErr:
		if err != nil {
			runErr = fmt.Errorf("orderapp, grpc serve: %w", err)
		}
	}

	workerCancel()
	app.shutDown()
	return runErr
}
