package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	DBPort      string `mapstructure:"DB_PORT"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
	ServerPort  string `mapstructure:"SERVER_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`

	// Настройки планировщика
	SchedulerEnabled   bool          `mapstructure:"SCHEDULER_ENABLED"`
	PostTickerDuration time.Duration `mapstructure:"POST_TICKER_DURATION"`
	PostWorkersCount   int           `mapstructure:"POST_WORKERS_COUNT"`
	PostBatchSize      int           `mapstructure:"POST_BATCH_SIZE"`
}

func Load() *Config {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Парсим время
	tickerDuration, err := time.ParseDuration(GetEnv("POST_TICKER_DURATION", "30s"))
	if err != nil {
		log.Fatalf("POST_TICKER_DURATION invalid format (use 10s, 30s, 1m): %v", err)
	}

	// Парсим количество воркеров
	workersCount, err := strconv.Atoi(GetEnv("POST_WORKERS_COUNT", "5"))
	if err != nil || workersCount < 1 || workersCount > 20 {
		log.Fatal("POST_WORKERS_COUNT invalid (must be 1-20)")
	}

	// Парсим размер батча
	batchSize, err := strconv.Atoi(GetEnv("POST_BATCH_SIZE", "50"))
	if err != nil || batchSize < 10 || batchSize > 100 {
		log.Fatal("POST_BATCH_SIZE invalid (must be 10-100)")
	}

	// Создаём конфиг из переменных окружения
	cfg := &Config{
		DBHost:      GetEnv("DB_HOST", "localhost"),
		DBPort:      GetEnv("DB_PORT", "5434"),
		DBUser:      GetEnv("DB_USER", "postgres"),
		DBPassword:  GetEnv("DB_PASSWORD", "postgres"),
		DBName:      GetEnv("DB_NAME", "postgres"),
		JWTSecret:   GetEnv("JWT_SECRET", ""),
		ServerPort:  GetEnv("SERVER_PORT", "8080"),
		Environment: GetEnv("ENVIRONMENT", "development"),

		// Планировщик из .env
		PostTickerDuration: tickerDuration,
		PostWorkersCount:   workersCount,
		PostBatchSize:      batchSize,
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

	log.Printf("📅 Scheduler config: ticker=%v, workers=%d, batch=%d",
		cfg.PostTickerDuration, cfg.PostWorkersCount, cfg.PostBatchSize)

	return cfg
}

// DatabaseURL формирует строку подключения PostgreSQL
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
