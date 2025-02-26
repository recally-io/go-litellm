package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/recally-io/polyllm"
)

func StartServer() {
	mux := http.NewServeMux()
	llmService := NewLLMService(polyllm.New())
	mux.HandleFunc("POST /chat/completions", authMiddleware(llmService.chatCompletion))
	mux.HandleFunc("POST /v1/chat/completions", authMiddleware(llmService.chatCompletion))

	mux.HandleFunc("GET /models", authMiddleware(llmService.listModels))
	mux.HandleFunc("GET /v1/models", authMiddleware(llmService.listModels))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
	slog.Info("Starting Litellm server", "port", port)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server", "err", err)
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
