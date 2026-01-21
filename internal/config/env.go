package config

import "os"

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	Port        string `env:"PORT,required"`
	JWTSecret   string `env:"JWT_SECRET,required"`
	Environment string `env:"ENVIRONMENT"`
}

func Load() *Config {
	cfg := &Config{}

	// Обязательные переменные
	cfg.DatabaseURL = GetEnv("DATABASE_URL", "postgres://user:pass@localhost/blogdb?sslmode=disable")
	cfg.Port = GetEnv("PORT", "8080")
	cfg.JWTSecret = GetEnv("JWT_SECRET", "super-secret-jwt-key-must-be-32-characters")
	cfg.Environment = GetEnv("ENVIRONMENT", "development")

	return cfg
}
