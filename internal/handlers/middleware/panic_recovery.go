package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// PanicRecoveryMiddleware — перехватывает panics
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v\n%s", err, debug.Stack())

				safeWriter := &safeStatusWriter{ResponseWriter: w, written: false}
				safeWriter.Header().Set("Content-Type", "application/json")
				safeWriter.WriteHeader(http.StatusInternalServerError)

				json.NewEncoder(safeWriter).Encode(ErrorResponse{
					Error: "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// safeStatusWriter — предотвращает повторную запись заголовков
type safeStatusWriter struct {
	http.ResponseWriter
	written bool
}

func (s *safeStatusWriter) WriteHeader(code int) {
	if s.written {
		return
	}
	s.written = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *safeStatusWriter) Header() http.Header {
	return s.ResponseWriter.Header()
}

func (s *safeStatusWriter) Write(b []byte) (int, error) {
	return s.ResponseWriter.Write(b)
}
