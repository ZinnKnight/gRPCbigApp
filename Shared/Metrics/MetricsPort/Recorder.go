package MetricsPort

type MetricsRecord interface {
	IncRequest(method, statusCode string)
}
