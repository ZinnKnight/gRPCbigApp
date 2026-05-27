package Tracing

import (
	"context"
	"fmt"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Logger         LoggerPorts.Logger
	ServiceName    string
	ServiceVersion string
	Environment    string
	SampleRatio    float64
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

func Init(ctx context.Context, config Config, exporter sdktrace.SpanExporter) (ShutDownTracing, error) {
	logger := config.Logger

	if exporter == nil {
		otel.SetTextMapPropagator(defaultPropagator())
		logInfo(logger, "tracing disabled, propagator only",
			field("service.name", config.ServiceName))
		return func(context.Context) error { return nil }, nil
	}

	if config.ServiceName == "" {
		return nil, fmt.Errorf("no tracing service name provided")
	}
	if logger != nil {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			logger.LogError("otel sdk runtime error", LoggerPorts.Field{Key: "error", Value: err.Error()})
		}))
	}

	res, err := buildResources(ctx, config)
	if err != nil {
		logErr(logger, "tracing: failed to build resorses", err)
		return nil, fmt.Errorf("failed to build resorses: %w", err)
	}

	// Formulating spans in butches and send like that, so its will cost less

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
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
	)

	return func(shutCTX context.Context) error {
		flushCTX, flushCancel := context.WithTimeout(shutCTX, time.Second*3)
		defer flushCancel()
		if err := traceProvider.ForceFlush(flushCTX); err != nil {
			logErr(logger, "tracing: failed to shutdown tracing", err)
		}
		if err := traceProvider.Shutdown(shutCTX); err != nil {
			logErr(logger, "tracing: failed to shutdown tracing", err)
			return err
		}
		logInfo(logger, "tracing: shutdown completed")
		return nil
	}, nil
}

func field(k string, v interface{}) LoggerPorts.Field {
	return LoggerPorts.Field{Key: k, Value: v}
}

func logInfo(l LoggerPorts.Logger, msg string, fields ...LoggerPorts.Field) {
	if l == nil {
		return
	}
	l.LogInfo(msg, fields...)
}

func logErr(l LoggerPorts.Logger, msg string, err error, extra ...LoggerPorts.Field) {
	if l == nil {
		return
	}
	all := append([]LoggerPorts.Field{{Key: "error", Value: err.Error()}}, extra...)
	l.LogError(msg, all...)
}
