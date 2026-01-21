package postgres

import (
	"blog-backend/internal/config"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Парамемтры подключений
	db.SetMaxOpenConns(25) // Максимум открытых соединений
	db.SetMaxIdleConns(25) // Максимум idle соединений
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// CloseDB закрывает соединение с базой данных
// func CloseDB() {
// 	if db != nil {
// 		db.Close()
// 	}
// }
