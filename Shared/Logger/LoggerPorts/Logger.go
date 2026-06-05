package LoggerPorts

type Logger interface {
	LogError(msg string, fields ...Field)
	LogInfo(msg string, fields ...Field)
	Sync() error
}

type Field struct {
	Key   string
	Value interface{}
}
