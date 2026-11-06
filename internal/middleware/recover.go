package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				slog.Error("panic",
					"recovery", rec,
					"stack", stack,
				)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
