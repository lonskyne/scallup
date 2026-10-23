package api

import (
	"net/http"

	"github.com/lonskyne/scallup/internal/config"
	"github.com/lonskyne/scallup/internal/raft"
	"github.com/lonskyne/scallup/internal/storage"
)

func SetupRoutes(storage storage.Engine, raftNode *raft.RaftNode, cfg *config.Config) http.Handler {
	handlers := NewHandlers(storage, raftNode, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/key/{key}", handlers.GetKey)
	mux.HandleFunc("POST /api/v1/key/{key}", handlers.PutKey)
	mux.HandleFunc("DELETE /api/v1/key/{key}", handlers.DeleteKey)
	mux.HandleFunc("GET /api/v1/keys", handlers.GetAllKeys)

	mux.HandleFunc("GET /health", handlers.HealthCheck)

	mux.HandleFunc("POST /pause", handlers.PauseRaftNode)
	mux.HandleFunc("POST /resume", handlers.ResumeRaftNode)

	return LoggingMiddleware(mux)
}
