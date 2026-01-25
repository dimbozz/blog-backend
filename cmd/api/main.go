package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/handlers"
	"blog-backend/internal/repository/postgres"
	"blog-backend/pkg/jwt"
	"blog-backend/service"
	"log"
	"net/http"
)

func main() {
	// Загрузка переменных окружения из .env файла
	cfg := config.Load()

	// Инициализация JWT секретного ключа
	jwt.InitAuth()

	// Инициализация БД с настройками пула (db.go)
	db, err := postgres.NewDB(cfg) // ← из db.go!
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаём слои снизу вверх (Repository → Service → Handler)
	// Каждый слой зависит только от интерфейса предыдущего

	// 1. Repository - уровень доступа к БД (конкретная реализация postgres)
	userRepo := postgres.NewPostgresUserRepository(db)

	// 2. Service - уровень бизнес-логики (зависит от интерфейса Repository)
	userService := service.NewUserService(userRepo)

	// 3. Handler - уровень HTTP (зависит от Service)
	userHandler := handlers.NewUserHandler(userService)

	// Настройка HTTP маршрутов
	http.HandleFunc("/api/register", userHandler.RegisterHandler)
	http.HandleFunc("/api/login", userHandler.LoginHandler)
	// http.HandleFunc("api/profile", middleware.AuthMiddleware(handlers.ProfileHandler(userRepo)))
	http.HandleFunc("/api/health", handlers.HealthHandler(userRepo))

	// Запуск сервера
	port := config.GetEnv("SERVER_PORT", "8080")
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📝 Register: POST http://localhost:%s/api/register", port)
	log.Printf("🔐 Login: POST http://localhost:%s/api/login", port)
	log.Printf("👤 Profile: GET http://localhost:%s/api/profile (requires token)", port)
	log.Printf("❤️  Health: GET http://localhost:%s/api/health", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
