package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/repository"
	"github.com/anddsdev/cloudlet/internal/server"
	"github.com/anddsdev/cloudlet/internal/services"
)

func main() {
	// Load configuration from environment variables first, fallback to YAML
	cfg, err := config.NewConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Log configuration source for debugging
	log.Printf("Configuration loaded successfully")
	log.Printf("Server will start on port: %s", cfg.Server.Port)
	log.Printf("Storage path: %s", cfg.Server.Storage.Path)
	log.Printf("Database DSN: %s", cfg.Database.DSN)

	repo, err := repository.NewFileRepository(cfg.Database.DSN, cfg.Database.MaxConn)
	if err != nil {
		log.Fatalf("error creating repository: %v", err)
	}
	defer repo.Close()

	// Verify that the storage directory exists
	if err := ensureStorageDirectory(cfg.Server.Storage.Path); err != nil {
		log.Fatalf("error ensuring storage directory: %v", err)
	}

	storageService := services.NewStorageService(cfg.Server.Storage.Path)
	defer storageService.Close()

	fileService := services.NewFileService(repo, storageService, cfg.Server.Storage.Path)

	httpServer := server.NewServer(cfg, fileService)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:           httpServer.Handler(),
		ReadTimeout:       time.Duration(cfg.Server.Timeout.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.Timeout.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.Timeout.IdleTimeout) * time.Second,
		MaxHeaderBytes:    cfg.Server.Timeout.MaxHeaderBytes,
		ReadHeaderTimeout: time.Duration(cfg.Server.Timeout.ReadHeaderTimeout) * time.Second,
	}

	go func() {
		log.Printf("starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited cleanly")
}

// Verify if the storage directory exists, and create it if not
func ensureStorageDirectory(storagePath string) error {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory %s: %w", storagePath, err)
		}
		log.Printf("Created storage directory: %s", storagePath)
	}
	return nil
}
