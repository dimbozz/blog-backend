package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Центральная обработка ошибок обработчиков (handlers)
func AbortError(w http.ResponseWriter, r *http.Request, msg string, status int, err error) {
	start := time.Now()

	// Единообразный лог (как в LoggingMiddleware)
	log.Printf("%s %s ERROR %d: %s (%v) %s %v",
		r.Method, r.URL.Path, status, msg, err, r.RemoteAddr, time.Since(start))

	// Ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
