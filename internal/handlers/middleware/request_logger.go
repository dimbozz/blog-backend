package middleware

import (
	"log"
	"net/http"
	"time"
)

// loggingResponseWriter для перехвата статуса ответа
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware — логирует запросы и ответы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &LoggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %s %d %v", r.Method, r.URL.Path, r.RemoteAddr, lrw.statusCode, duration)
	})
}
