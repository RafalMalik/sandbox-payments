package api

import (
	"log/slog"
	"net/http"
	"time"
)

// loggingMiddleware logs each HTTP request with method, path and duration.
func loggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

// recoverMiddleware converts panics into 500 responses.
func recoverMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "error", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NewServer builds the root HTTP handler with middleware and static assets.
func NewServer(handler *Handler, staticDir string, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	mux.Handle("/logos/", http.StripPrefix("/logos/", http.FileServer(http.Dir(staticDir+"/logos"))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	var root http.Handler = mux
	root = loggingMiddleware(log, root)
	root = recoverMiddleware(log, root)
	return root
}
