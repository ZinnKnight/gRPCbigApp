package Tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ propagation.TextMapCarrier = (TraceCarier)(nil)

type TraceCarier map[string]string

func (car TraceCarier) Get(key string) string {
	return car[key]
}

func (car TraceCarier) Set(key string, value string) {
	car[key] = value
}

func (car TraceCarier) Keys() []string {
	keys := make([]string, 0, len(car))
	for k := range car {
		keys = append(keys, k)
	}
	return keys
}

func PlaceIntoCar(ctx context.Context) TraceCarier {
	car := TraceCarier{}
	otel.GetTextMapPropagator().Inject(ctx, car)
	return car
}

func TakeOutFromCar(ctx context.Context, car TraceCarier) context.Context {
	if len(car) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, car)
}
