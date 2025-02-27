package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/logger"
)

func StartServer() {
	mux := http.NewServeMux()
	llmService := NewLLMService(polyllm.New())
	mux.HandleFunc("POST /chat/completions", loggingMiddleware(authMiddleware(llmService.chatCompletion)))
	mux.HandleFunc("POST /v1/chat/completions", loggingMiddleware(authMiddleware(llmService.chatCompletion)))

	mux.HandleFunc("GET /models", loggingMiddleware(authMiddleware(llmService.listModels)))
	mux.HandleFunc("GET /v1/models", loggingMiddleware(authMiddleware(llmService.listModels)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
	slog.Info("Starting polyllm server", "port", port)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server", "err", err)
	}
}

// loggingMiddleware logs information about each incoming request
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture the status code
		rw := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(rw, r)

		// Log request details after processing
		duration := time.Since(start)
		logger.DefaultLogger.Info("Request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	}
}

// responseWriter is a custom wrapper to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush implements the http.Flusher interface
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		if apiKey != "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			headerKey := strings.TrimPrefix(authHeader, "Bearer ")
			if headerKey != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}
