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
	cfg, err := config.NewConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	repo, err := repository.NewFileRepository(cfg.Database.DSN, cfg.Database.MaxConn)
	if err != nil {
		log.Fatalf("error creating repository: %v", err)
	}
	defer repo.Close()

	storageService := services.NewStorageService(cfg.Server.Storage.Path)
	defer storageService.Close() // Ensure proper cleanup of atomic operations
	
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
