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
	mux.HandleFunc("POST /api/v1/upload/stream", r.withMiddleware(h.UploadStream))
	mux.HandleFunc("POST /api/v1/upload/chunked", r.withMiddleware(h.UploadChunked))
	mux.HandleFunc("POST /api/v1/upload/progress", r.withMiddleware(h.UploadWithProgressTracking))
	mux.HandleFunc("GET /api/v1/download/{path}", r.withMiddleware(h.Download))

	// Directories operations
	mux.HandleFunc("POST /api/v1/directories", r.withMiddleware(h.CreateDirectory))
	mux.HandleFunc("GET /api/v1/directories/{path}", r.withMiddleware(h.ListFiles))

	// Operations on directories
	mux.HandleFunc("POST /api/v1/move", r.withMiddleware(h.MoveFile))
	mux.HandleFunc("POST /api/v1/rename", r.withMiddleware(h.RenameFile))

	fs := http.FileServer(http.Dir("./web/"))
	mux.Handle("/", r.withMiddleware(http.StripPrefix("/", fs).ServeHTTP))

	r.handler = mux
}

func (r *Router) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return r.cors(r.logging(r.recovery(next)))
}
