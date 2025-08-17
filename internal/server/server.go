package server

import (
	"net/http"

	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/services"
)

type Server struct {
	cfg         *config.Config
	router      *Router
	fileService *services.FileService
}

func NewServer(cfg *config.Config, fileService *services.FileService) *Server {
	s := &Server{
		cfg:         cfg,
		fileService: fileService,
	}

	s.router = NewRouter(s)

	return s
}

func (s *Server) Handler() http.Handler {
	return s.router.handler
}

func (s *Server) Config() *config.Config {
	return s.cfg
}

func (s *Server) FileService() *services.FileService {
	return s.fileService
}
