package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type Config struct {
	DBHost      string `mapstructure:"DB_HOST"`
	DBPort      int    `mapstructure:"DB_PORT"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
	ServerPort  string `mapstructure:"SERVER_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
}

func Load() *Config {
	// Загружаем .env
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("No .env file found")
	// }

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// ЧИТАЕМ ФАЙЛ
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading .env: %v", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Валидация
	if cfg.DBHost == "" || cfg.DBName == "" || cfg.DBUser == "" {
		log.Fatal("DB_HOST, DB_NAME, DB_USER required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET required (min 32 chars)")
	}
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET too short (min 32 chars)")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}

	return cfg
}

// DatabaseURL формирует строку подключения PostgreSQL
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
