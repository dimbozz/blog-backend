package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

// Центральная обработка ошибок!
func AbortError(w http.ResponseWriter, r *http.Request, msg string, status int, err error) {
	// Логируем ошибку
	log.Printf("%s %s ERROR %d: %s (%v)", r.Method, r.URL.Path, status, msg, err)

	// Отвечаем клиенту (safeStatusWriter уже есть в panic_recovery.go)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
