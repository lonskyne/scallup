package api

import (
    "net/http"

		"github.com/lonskyne/scallup/internal/storage"
)

func SetupRoutes(storage storage.Engine) http.Handler {
    handlers := NewHandlers(storage)
    
    mux := http.NewServeMux()
    
    mux.HandleFunc("GET /api/v1/key/{key}", handlers.GetKey)
    mux.HandleFunc("POST /api/v1/key/{key}", handlers.PutKey)
    mux.HandleFunc("DELETE /api/v1/key/{key}", handlers.DeleteKey)
    mux.HandleFunc("GET /api/v1/keys", handlers.GetAllKeys)
    
    mux.HandleFunc("GET /health", handlers.HealthCheck)
    
    return LoggingMiddleware(mux)
}
