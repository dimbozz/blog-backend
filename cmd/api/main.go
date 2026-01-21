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

	cfg := config.Load()

	// Инициализация JWT секретного ключа
	jwt.InitAuth()

	// Инициализация БД с настройками пула (db.go)
	db, err := postgres.NewDB(cfg) // ← из db.go!
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем экземпляр репозитория (user.go)
	userRepo := postgres.NewPostgresUserRepository(db)

	// TODO: Настройка HTTP маршрутов
	// Используйте обработчики из handlers.go
	http.HandleFunc("/register", handlers.RegisterHandler(userRepo))
	http.HandleFunc("/login", handlers.LoginHandler(userRepo))
	http.HandleFunc("/profile", middleware.AuthMiddleware(handlers.ProfileHandler(userRepo)))
	http.HandleFunc("/health", handlers.HealthHandler(userRepo))

	// Запуск сервера
	port := config.GetEnv("SERVER_PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📝 Register: POST http://localhost:%s/register", port)
	log.Printf("🔐 Login: POST http://localhost:%s/login", port)
	log.Printf("👤 Profile: GET http://localhost:%s/profile (requires token)", port)
	log.Printf("❤️  Health: GET http://localhost:%s/health", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
