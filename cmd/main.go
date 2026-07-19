package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/lonskyne/scallup/internal/api"
	"github.com/lonskyne/scallup/internal/config"
	"github.com/lonskyne/scallup/internal/storage"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize storage
    store := createStore(cfg)
		defer store.Close()
    
    // Setup routes
    router := api.SetupRoutes(store)
    
    // Create server
    server := &http.Server{
        Addr:         cfg.APIAddr(),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        log.Printf("Starting scallup instance API on %s", cfg.APIAddr())
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Server stopped")
}

func createStore(cfg *config.Config) storage.Engine {
	switch(cfg.DBType) {
	case "memory":
		return storage.NewMemoryStore()
	case "jsonfile":
		return storage.NewJSONFileStore(filepath.Join(cfg.DataDir, cfg.DBName + ".json"))
	}

	return storage.NewMemoryStore()
}
