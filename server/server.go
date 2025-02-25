package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

func StartServer() {
	if err := initProviders(); err != nil {
		slog.Error("failed to init providers", "err", err)
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat/completions", authMiddleware(completionHandler))
	mux.HandleFunc("POST /v1/chat/completions", authMiddleware(completionHandler))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", Settings.PORT),
		Handler: mux,
	}
	slog.Info("Starting Litellm server", "port", Settings.PORT)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server", "err", err)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Add token validation logic here
		next.ServeHTTP(w, r)
	}
}
