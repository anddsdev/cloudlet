package server

import (
	"net/http"

	"github.com/anddsdev/cloudlet/internal/handlers"
)

type Router struct {
	server  *Server
	handler http.Handler
}

func NewRouter(server *Server) *Router {
	r := &Router{
		server: server,
	}

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	mux := http.NewServeMux()

	h := handlers.NewHandlers(r.server.FileService(), r.server.Config())

	mux.HandleFunc("GET /health", r.withMiddleware(h.HealthCheck))

	mux.HandleFunc("GET /api/v1/files", r.withMiddleware(h.ListFiles))
	mux.HandleFunc("GET /api/v1/files/{path}", r.withMiddleware(h.ListFiles))
	mux.HandleFunc("DELETE /api/v1/files", r.withMiddleware(h.DeleteFile))
	mux.HandleFunc("POST /api/v1/upload", r.withMiddleware(h.Upload))
	mux.HandleFunc("GET /api/v1/download/{path}", r.withMiddleware(h.Download))

	fs := http.FileServer(http.Dir("./web/"))
	mux.Handle("/", r.withMiddleware(http.StripPrefix("/", fs).ServeHTTP))

	r.handler = mux
}

func (r *Router) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return r.cors(r.logging(r.recovery(next)))
}
