package handlers

import (
	"blog-backend/internal/repository"
	"encoding/json"
	"log"
	"net/http"
)

// HealthHandler проверяет состояние сервиса
func HealthHandler(repo repository.HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path) // Логирование запроса
		// Проверяем подключение к БД
		if err := repo.HealthCheck(r.Context()); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		// Возвращаем статус OK
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"status":  "ok",
			"message": "Service is running",
		}
		json.NewEncoder(w).Encode(response)
	}
}
