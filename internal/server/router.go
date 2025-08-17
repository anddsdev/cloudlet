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

	fs := http.FileServer(http.Dir("./web/"))
	mux.Handle("/", r.withMiddleware(http.StripPrefix("/", fs).ServeHTTP))

	r.handler = mux
}

func (r *Router) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return r.cors(r.logging(r.recovery(next)))
}
