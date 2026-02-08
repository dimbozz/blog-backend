package main

import (
	"blog-backend/internal/config"
	"blog-backend/internal/handlers"
	"blog-backend/internal/handlers/middleware"
	"blog-backend/internal/repository/postgres"
	"blog-backend/pkg/jwt"
	"blog-backend/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// 1. Создаем mux
	mux := http.NewServeMux()
	// Создаём слои снизу вверх (Repository → Service → Handler)
	// Каждый слой зависит только от интерфейса предыдущего

	// Repository - уровень доступа к БД (конкретная реализация postgres)
	userRepo := postgres.NewPostgresUserRepository(db)
	postRepo := postgres.NewPostgresPostRepository(db)
	commentRepo := postgres.NewPostgresCommentRepository(db)

	// Service - уровень бизнес-логики (зависит от интерфейса Repository)
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo, cfg)
	commentService := service.NewCommentService(postRepo, commentRepo, userRepo)

	// Логгер
	stdLogger := log.New(log.Writer(), "", log.LstdFlags)

	// Handler - уровень HTTP (зависит от Service)
	userHandler := handlers.NewUserHandler(userService, stdLogger)
	postHandler := handlers.NewPostHandler(postService, stdLogger)
	commentHandler := handlers.NewCommentHandler(commentService)

	// Настройка HTTP маршрутов для пользователей
	mux.HandleFunc("/api/register", userHandler.RegisterHandler)
	mux.HandleFunc("/api/login", userHandler.LoginHandler)
	mux.HandleFunc("/api/profile", middleware.AuthMiddleware(userHandler.ProfileHandler))
	mux.HandleFunc("/api/health", handlers.HealthHandler(userRepo))

	// Настройка HTTP маршрутов для постов
	// GET /api/posts — получить список постов (доступно всем)
	// POST /api/posts — создать пост (только авторизованный пользователь)
	mux.HandleFunc("GET /api/posts", postHandler.ListPosts)
	mux.HandleFunc("POST /api/posts", middleware.AuthMiddleware(postHandler.CreatePost))

	// GET /api/posts/{postid} — получить один пост
	// PUT /api/posts/{postid} — обновить пост (только автор)
	// DELETE /api/posts/{postid} — удалить пост (только автор)
	mux.HandleFunc("GET /api/posts/{postid}", postHandler.GetPost)
	mux.HandleFunc("PUT /api/posts/{postid}", middleware.AuthMiddleware(postHandler.UpdatePost))
	mux.HandleFunc("DELETE /api/posts/{postid}", middleware.AuthMiddleware(postHandler.DeletePost))

	// Настройка HTTP маршрутов для комментариев
	mux.HandleFunc("POST /api/posts/{postId}/comments", middleware.AuthMiddleware(commentHandler.CreateComment))
	mux.HandleFunc("GET /api/posts/{postId}/comments", commentHandler.GetComments)

	// 2. Оборачиваем mux в middleware цепочку
	handler := middleware.LoggingMiddleware(mux)
	handler = middleware.PanicRecoveryMiddleware(handler)

	// Создаем HTTP сервер для graceful shutdown
	port := config.GetEnv("SERVER_PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler, // mux с middleware
		// Graceful shutdown настройки
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Канал для сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запускаем отдельную горутину с сервером
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		log.Printf("📝 Register: POST http://localhost:%s/api/register", port)
		log.Printf("🔐 Login: POST http://localhost:%s/api/login", port)
		log.Printf("👤 Profile: GET http://localhost:%s/api/profile (requires token)", port)
		log.Printf("❤️ Health: GET http://localhost:%s/api/health", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 6. Ждем сигнал завершения
	<-quit
	log.Println("🛑 Shutdown signal received, starting graceful shutdown...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем планировщик
	go func() {
		log.Println("Stopping post scheduler...")
		postService.Stop()
	}()

	// Останавливаем HTTP сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server forced shutdown: %v", err)
	} else {
		log.Println("HTTP Server stopped")
	}

	// Закрываем БД соединения
	db.SetMaxOpenConns(0)

	log.Println("✅ Graceful shutdown complete!")
}
