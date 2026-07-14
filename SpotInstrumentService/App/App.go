package App

import (
	"context"
	"fmt"
	authInterceptor "gRPCbigapp/ClientService/Auth/AuthInterceptor"
	"gRPCbigapp/OrderService/Txmanager"
	"gRPCbigapp/Proto/protoPB"
	"gRPCbigapp/Shared/Config"
	"gRPCbigapp/Shared/ErrorInterceptor"
	"gRPCbigapp/Shared/Events"
	"gRPCbigapp/Shared/Idempotentor"
	"gRPCbigapp/Shared/Kafka"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	Metrics2 "gRPCbigapp/Shared/Metrics"
	"gRPCbigapp/Shared/Outbox"
	"gRPCbigapp/Shared/PanicInterceptor"
	pgPool "gRPCbigapp/Shared/Postgres"
	RateLimiter2 "gRPCbigapp/Shared/RateLimiter"
	redisClient "gRPCbigapp/Shared/Redis"
	"gRPCbigapp/Shared/SagaMessages"
	tracing "gRPCbigapp/Shared/Tracing"
	otlpexporter "gRPCbigapp/Shared/Tracing/OPTLExp"
	"gRPCbigapp/Shared/ValidationIntercepter"
	marketPG "gRPCbigapp/SpotInstrumentService/Adapters/Postgres"
	marketGRPC "gRPCbigapp/SpotInstrumentService/Adapters/grpcAdapter"
	"gRPCbigapp/SpotInstrumentService/SagaSubs"
	marketUC "gRPCbigapp/SpotInstrumentService/UseCase"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

const suhtdownTimer = 15 * time.Second

type App struct {
	cfg    *Config.Config
	logger LoggerPorts.Logger

	pool          *pgxpool.Pool
	rdb           *redisClient.RedisClient
	producer      *Kafka.Producer
	relay         *Outbox.Relay
	cmdConsumer   *Kafka.Consumer
	subscriber    *SagaSubs.SagaSub
	grpcServer    *grpc.Server
	listen        net.Listener
	metricsHandle http.Handler

	tracingShutDown tracing.ShutDownTracing
}

func New(ctx context.Context, cfg *Config.Config, logger LoggerPorts.Logger) (*App, error) {
	app := &App{
		cfg:    cfg,
		logger: logger,
	}
	var tracerExp sdktrace.SpanExporter
	if cfg.TracingEnabled {
		exp, err := otlpexporter.NewgRPCExporter(ctx, cfg.OpenTelemetryEndpoint)
		if err != nil {
			return nil, fmt.Errorf("market.otlp exporter: %w", err)
		}
		tracerExp = exp
	}

	tracingShutDown, err := tracing.Init(ctx, tracing.Config{
		Logger:         logger,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		SampleRatio:    cfg.TracingSampleRatio,
	}, tracerExp)
	if err != nil {
		return nil, fmt.Errorf("marketApp, tracing.init: %w", err)
	}
	app.tracingShutDown = tracingShutDown

	pool, err := pgPool.NewPool(ctx, pgPool.Config{
		DatabaseURL:    cfg.DatabaseURL,
		MaxConnections: int32(cfg.DBMaxConn),
		MinConnections: int32(cfg.DBMinConn),
		MaxConnTTL:     time.Duration(cfg.DBMaxConnTTL) * time.Minute,
		DBMaxConnIdTTL: time.Duration(cfg.DBMaxConnIdTTL) * time.Minute,
	})
	if err != nil {
		_ = app.tracingShutDown(ctx)
		return nil, fmt.Errorf("marketApp, pool.New: %w", err)
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
		_ = app.tracingShutDown(ctx)
		return nil, fmt.Errorf("marketApp, rdb.New: %w", err)
	}
	app.rdb = rdb

	producer, err := Kafka.NewProducer(ctx, Kafka.Config{
		Brokers:  cfg.KafkaBrokers,
		ClientID: cfg.ServiceName,
	})
	if err != nil {
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutDown(ctx)
		return nil, fmt.Errorf("marketApp, producer.New: %w", err)
	}
	app.producer = producer

	txManager := Txmanager.NewTxManager(pool)

	marketRepo := marketPG.NewSISMarketRepo(pool)
	merketUseCase := marketUC.NewSISUseCase(marketRepo, logger)
	marketHandler := marketGRPC.NewSISgrpcHandler(merketUseCase, logger)

	replyResolve := func(e Events.Events) string {
		switch e.EventType {
		case SagaMessages.EventStockReserved, SagaMessages.EventStockRejected:
			return SagaMessages.TopicSagaReplies
		default:
			return e.AggregationType + ".events"
		}
	}
	replyWriter := Outbox.NewWriter(pool, replyResolve)
	app.relay = Outbox.NewRelay(pool, producer, logger, 100, time.Second)

	app.subscriber = SagaSubs.NewSagaSub(marketRepo, txManager, replyWriter, Idempotentor.NewGuard(pool, "market-reserv"), logger)

	cmdConsume, err := Kafka.NewConsumer(ctx, Kafka.ConsumerConfig{
		Brokers: cfg.KafkaBrokers,
		Group:   "market-reserve",
		Topics:  []string{SagaMessages.TopicSagaCommands},
	}, logger)
	if err != nil {
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutDown(ctx)
		return nil, fmt.Errorf("marketApp, cmdConsume: %w", err)
	}
	app.cmdConsumer = cmdConsume

	limit := RateLimiter2.NewRateLimiter(rdb.Client, cfg.RateLimitPerMin, time.Minute)

	metricsRecord := Metrics2.NewPrometheusRecord()
	app.metricsHandle = metricsRecord.Registry()

	app.grpcServer = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			PanicInterceptor.PanicRecoveryInterceptor(logger),
			Metrics2.UnaryServerInterceptor(metricsRecord),
			ErrorInterceptor.UnaryServerInterceptor(logger),
			ValidationIntercepter.UnaryServerInterceptor(),
			authInterceptor.AuthInterceptor([]byte(cfg.JWTSecretKey)),
			RateLimiter2.UnaryServerInterceptor(limit),
		),
	)
	protoPB.RegisterSpotInstrumentServiceServer(app.grpcServer, marketHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		app.cmdConsumer.Close()
		app.producer.Close()
		app.rdb.Close()
		app.pool.Close()
		_ = app.tracingShutDown(ctx)
		return nil, fmt.Errorf("marketApp, net.Listen: %w", err)
	}
	app.listen = lis

	return app, nil
}

func (app *App) shutDown() {
	stopChan := make(chan struct{})

	go func() {
		app.grpcServer.Stop()
		close(stopChan)
	}()
	select {
	case <-stopChan:
		app.logger.LogInfo("marketApp shutting down")
	case <-time.After(suhtdownTimer):
		app.logger.LogError("marketApp forced shutting down")
		app.grpcServer.Stop()
		<-stopChan
	}

	if app.cmdConsumer != nil {
		app.cmdConsumer.Close()
	}
	if app.producer != nil {
		app.producer.Close()
	}

	app.pool.Close()
	if err := app.rdb.Close(); err != nil {
		app.logger.LogError("marketApp, redis stoped", LoggerPorts.Field{Key: "error", Value: err.Error()})
	}

	shutDownCTX, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.tracingShutDown(shutDownCTX); err != nil {
		app.logger.LogError("marketApp shutting down", LoggerPorts.Field{Key: "error", Value: err.Error()})
	}
}

func (app *App) Run(ctx context.Context) error {
	workerCtx, workerCancel := context.WithCancel(ctx)

	go func() {
		if err := Metrics2.StartMetricsServer(workerCtx, app.cfg.MetricsPort, app.metricsHandle); err != nil {
			app.logger.LogError("marketApp metrics", LoggerPorts.Field{Key: "error", Value: err.Error()})
		}
	}()

	go app.relay.Run(workerCtx)
	go app.cmdConsumer.Run(workerCtx, app.subscriber.HandleReservedStock)

	servErrChan := make(chan error, 1)

	go func() {
		app.logger.LogInfo("market-service listening",
			LoggerPorts.Field{Key: "grpc_port", Value: app.cfg.GRPCPort},
			LoggerPorts.Field{Key: "metrics", Value: app.metricsHandle},
		)
		if err := app.grpcServer.Serve(app.listen); err != nil {
			servErrChan <- err
			return
		}
		servErrChan <- nil
	}()

	var runErr error

	select {
	case <-ctx.Done():
		app.logger.LogInfo("marketApp shutting down")
	case err := <-servErrChan:
		if err != nil {
			runErr = fmt.Errorf("marketApp, grpc serve: %w", err)
		}
	}
	workerCancel()
	app.shutDown()
	return runErr
}
