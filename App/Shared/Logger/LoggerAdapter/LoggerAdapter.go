package LoggerAdapter

import (
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"

	"go.uber.org/zap"
)

type ZapLogger struct {
	log *zap.Logger
}

func NewZapLogger() (*ZapLogger, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, err // Not sure can i log a process of making a logger?
	}
	return &ZapLogger{log: l}, nil
}

func toZapLogger(fields []LoggerPorts.Fieled) []zap.Field {
	res := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		res = append(res, zap.Any(field.Key, field.Value))
	}
	return res
}

func (l *ZapLogger) LogError(msg string, fields ...LoggerPorts.Fieled) {
	l.log.Error(msg, toZapLogger(fields)...)
}

func (l *ZapLogger) LogInfo(msg string, fields ...LoggerPorts.Fieled) {
	l.log.Info(msg, toZapLogger(fields)...)
}

func (l *ZapLogger) Sync() error {
	return l.log.Sync()
}
