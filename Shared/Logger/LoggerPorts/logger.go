package LoggerPorts

type Logger interface {
	LogError(msg string, fields ...Fieled)
	LogInfo(msg string, fields ...Fieled)
}

type Fieled struct {
	Key   string
	Value interface{}
}
