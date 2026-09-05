package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/lonskyne/scallup/internal/api"
	"github.com/lonskyne/scallup/internal/config"
	"github.com/lonskyne/scallup/internal/raft"
	"github.com/lonskyne/scallup/internal/storage"
	"github.com/lonskyne/scallup/pkg/pb"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize storage
	store, err := createStore(cfg)
	if err != nil {
		log.Printf("failed creating db engine: %s", err)
		return
	}
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

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterRaftServer(grpcServer, raft.NewRaftServer())

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.RaftGrpcPort))
	if err != nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}

	go func() {
		log.Printf("Starting grpc server on %s", cfg.RaftAddr())
		if serveErr := grpcServer.Serve(listener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			log.Printf("gRPC server stopped: %v", serveErr)
		}
	}()


	wal := store.GetWAL()
	raftStorageFilePath := filepath.Join(cfg.DataDir, cfg.DBName+".raft")
	node, err := raft.NewRaftNode(cfg.NodeID, cfg.Peers, raftStorageFilePath, wal)
	if err != nil {
		log.Fatalf("Node creation failed %v", err)
	}
	node.Initialize(context.Background())

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

	log.Println("Shutting down gRPC server...")

	grpcServer.Stop()

	log.Println("gRPC server stopped")
}

func createStore(cfg *config.Config) (storage.Engine, error) {
	walFilePath := filepath.Join(cfg.DataDir, cfg.DBName+".wal")
	switch cfg.DBType {
	case "memory":
		return storage.NewMemoryStore(walFilePath)
	case "jsonfile":
		return storage.NewJSONFileStore(filepath.Join(cfg.DataDir, cfg.DBName+".json"), walFilePath)
	}

	return storage.NewMemoryStore(walFilePath)
}
