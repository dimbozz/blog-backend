package config

import "os"

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
