package handlers

import (
	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/services"
)

type Handlers struct {
	fileService *services.FileService
	cfg         *config.Config
}

func NewHandlers(fileService *services.FileService, cfg *config.Config) *Handlers {
	return &Handlers{
		fileService: fileService,
		cfg:         cfg,
	}
}
