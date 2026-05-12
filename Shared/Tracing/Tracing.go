package Tracing

import (
	"context"
	"errors"
	"fmt"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Logger         LoggerPorts.Logger
	Endpoint       string
	ServiceName    string
	ServiceVersion string
	Environment    string
	SampleRatio    float64
	Enabled        bool
}

func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func defaultPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func buildResources(ctx context.Context, config Config) (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		attribute.String("service.name", config.ServiceName),
	}

	if config.ServiceVersion != "" {
		attributes = append(attributes, attribute.String("service.version", config.ServiceVersion))
	}

	if config.Environment != "" {
		attributes = append(attributes, attribute.String("environment", config.Environment))
	}

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attributes...),
	)
}

//For call a shutdown without loose last traces, wich still can going. Basicly works as Gracefull Shutdown for trces

type ShutDownTracing func(context.Context) error

func Init(ctx context.Context, config Config) (ShutDownTracing, error) {
	logger := config.Logger

	if !config.Enabled {
		otel.SetTextMapPropagator(defaultPropagator())
		logInfo(logger, "tracing disabled, propagator only",
			field("service.name", config.ServiceName))
		return func(context.Context) error { return nil }, nil
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("no tracing endpoint provided")
	}
	if config.ServiceName == "" {
		return nil, fmt.Errorf("no tracing service name provided")
	}
	if logger != nil {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			logger.LogError("otel sdk runtime error", LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		}))
	}

	res, err := buildResources(ctx, config)
	if err != nil {
		logErr(logger, "tracing: failed to build resorses", err)
		return nil, fmt.Errorf("failed to build resorses: %w", err)
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.Endpoint),
		// I not fully get how to make a credetionals +, its whole system work on localhost so its not really needed
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		logErr(logger, "tracing: otlp exporter failed", err, field("endpoint", config.Endpoint))
		return nil, fmt.Errorf("failed to init OTLP tracing: %w", err)
	}

	// Formulating spans in butches and send like that, so its will cost less

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxQueueSize(2048),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(5*time.Second),
		),

		sdktrace.WithSampler(
			sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SampleRatio)),
		),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(defaultPropagator())

	logInfo(logger, "tracing: initialized",
		field("service.name", config.ServiceName),
		field("service.version", config.ServiceVersion),
		field("environment", config.Environment),
		field("sample_ratio", config.SampleRatio),
		field("endpoint", config.Endpoint),
	)

	return func(shutCTX context.Context) error {
		flushCTX, flushCancel := context.WithTimeout(shutCTX, time.Second*3)
		defer flushCancel()
		if err := traceProvider.ForceFlush(flushCTX); err != nil {
			logErr(logger, "tracing: failed to shutdown tracing", err)
		}

		shutDownErr := errors.Join(
			traceProvider.Shutdown(shutCTX),
			exp.Shutdown(shutCTX),
		)

		if shutDownErr != nil {
			logErr(logger, "tracing: failed to shutdown tracing", shutDownErr)
		} else {
			logInfo(logger, "tracing: shutdown succeeded")
		}
		return shutDownErr
	}, nil
}

func field(k string, v interface{}) LoggerPorts.Fieled {
	return LoggerPorts.Fieled{Key: k, Value: v}
}

func logInfo(l LoggerPorts.Logger, msg string, fields ...LoggerPorts.Fieled) {
	if l == nil {
		return
	}
	l.LogInfo(msg, fields...)
}

func logErr(l LoggerPorts.Logger, msg string, err error, extra ...LoggerPorts.Fieled) {
	if l == nil {
		return
	}
	all := append([]LoggerPorts.Fieled{{Key: "error", Value: err.Error()}}, extra...)
	l.LogError(msg, all...)
}
