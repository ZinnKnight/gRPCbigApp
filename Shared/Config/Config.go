package Config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
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
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gRPCbigApp"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		GRPCPort:        getEnvInt("GRPC_PORT", 50051),
		MetricsPort:     getEnvInt("MetricsPort", 2112),
		JWTSecretKey:    getEnv("JWT_SECRET", ""),
		OutBoxInterval:  getEnvInt("OUTBOX_POLL_INTERVAL", 5),
		OutBoxButchSize: getEnvInt("OUTBOX_BUTCH_SIZE", 10),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 100),
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
