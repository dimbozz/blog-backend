package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/handlers"
	"blog-backend/internal/handlers/middleware"
	"blog-backend/internal/repository/postgres"
	"blog-backend/pkg/jwt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Инициализация JWT секретного ключа
	jwt.InitAuth()

	// TODO: Инициализация подключения к базе данных
	// Используйте функцию InitDB() из database.go
	if err := postgres.InitDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer postgres.CloseDB()

	// TODO: Настройка HTTP маршрутов
	// Используйте обработчики из handlers.go
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/profile", middleware.AuthMiddleware(handlers.ProfileHandler))
	http.HandleFunc("/health", handlers.HealthHandler)

	// Запуск сервера
	port := config.GetEnv("SERVER_PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📝 Register: POST http://localhost:%s/register", port)
	log.Printf("🔐 Login: POST http://localhost:%s/login", port)
	log.Printf("👤 Profile: GET http://localhost:%s/profile (requires token)", port)
	log.Printf("❤️  Health: GET http://localhost:%s/health", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
