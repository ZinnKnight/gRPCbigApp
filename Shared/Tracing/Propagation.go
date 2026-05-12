package Tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ propagation.TextMapCarrier = (TraceCarrier)(nil)

type TraceCarrier map[string]string

func (car TraceCarrier) Get(key string) string {
	return car[key]
}

func (car TraceCarrier) Set(key string, value string) {
	car[key] = value
}

func (car TraceCarrier) Keys() []string {
	keys := make([]string, 0, len(car))
	for k := range car {
		keys = append(keys, k)
	}
	return keys
}

func PlaceIntoCar(ctx context.Context) TraceCarrier {
	car := TraceCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, car)
	return car
}

func TakeOutFromCar(ctx context.Context, car TraceCarrier) context.Context {
	if len(car) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, car)
}
