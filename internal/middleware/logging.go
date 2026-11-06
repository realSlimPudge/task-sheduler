package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := NewResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		slog.Info("http request",
			"requestID", GetRequestID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"code", wrapped.statusCode,
			"duration", duration,
		)
	})
}
