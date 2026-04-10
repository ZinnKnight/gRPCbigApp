package Config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURl     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	GRPCPort        int
	MetricsPort     int
	JWTSecretKey    string
	OutBoxInterval  int
	OutBoxButchSize int
	RateLimitPerMin int
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		// TODO normal routing and combination with docker-compose + fallbacks
		DatabaseURl:     os.Getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gRPCbigApp"),
		RedisAddr:       os.Getenv("REDIS_URL", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD", ""),
		RedisDB:         os.Getenv("REDIS_DB", 0),
		GRPCPort:        os.Getenv("GRPC_PORT", 50051),
		MetricsPort:     os.Getenv("MetricsPort", 2112),
		JWTSecretKey:    os.Getenv("JWT_SECRET", ""),
		OutBoxInterval:  os.Getenv("OUTBOX_POLL_INTERVAL", 5),
		OutBoxButchSize: os.Getenv("OUTBOX_BUTCH_SIZE", 10),
		RateLimitPerMin: os.Getenv("RATE_LIMIT_PER_MIN", 100),
	}

	if cfg.JWTSecretKey == "" {
		return nil, fmt.Errorf("config, JWT_SECRET is required")
	}
	return cfg, nil
}

// nums for outbox quite random, need more examples of real code for figure outing

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
