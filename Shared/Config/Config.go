package Config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisPoolSize     int
	RedisMinIdleConns int
	GRPCPort          int
	MetricsPort       int
	JWTSecretKey      string
	RateLimitPerMin   int
	// for Postres pool (вынес отдельно сюда значения, что позже будут поднянуты в отдельном слое)

	DBMaxConn      int
	DBMinConn      int
	DBMaxConnTTL   int
	DBMaxConnIdTTL int

	// for Jaeger
	ServiceName           string
	ServiceVersion        string
	Environment           string
	OpenTelemetryEndpoint string
	TracingEnabled        bool
	TracingSampleRatio    float64
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gRPCbigApp"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		RedisPoolSize:   getEnvInt("REDIS_POOL_SIZE", 100),
		GRPCPort:        getEnvInt("GRPC_PORT", 50051),
		MetricsPort:     getEnvInt("MetricsPort", 2112),
		JWTSecretKey:    getEnv("JWT_SECRET", ""),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 100),
		// for Pool
		DBMaxConn:      getEnvInt("DB_MAX_CONN", 50),
		DBMinConn:      getEnvInt("DB_MIN_CONN", 10),
		DBMaxConnTTL:   getEnvInt("DB_MAX_CONN_TTL", 30),
		DBMaxConnIdTTL: getEnvInt("DB_MIN_CONN_TTL", 5),
		// for Jaeger
		ServiceName:           getEnv("SERVICE_NAME", "unknown service"),
		ServiceVersion:        getEnv("SERVICE_VERSION", "dev"),
		Environment:           getEnv("ENVIRONMENT", "local"),
		OpenTelemetryEndpoint: getEnv("OPEN_TELEMETRY_ENDPOINT", "jaeger:4317"), // default
		TracingEnabled:        getEnvBool("TRACING_ENABLED", true),
		TracingSampleRatio:    getEnvFloat("TRACING_SAMPLE_RATIO", 1.0),
	}

	if cfg.JWTSecretKey == "" {
		return nil, fmt.Errorf("config, JWT_SECRET is required")
	}
	return cfg, nil
}

// nums for outbox are standard, no big thoughts here

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
