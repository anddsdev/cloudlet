package server

import (
	"log"
	"net/http"
	"time"
)

func (r *Router) recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next(w, req)
	}
}

func (r *Router) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, req)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *Router) logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(ww, req)

		if ww.statusCode >= http.StatusBadRequest || req.Method != http.MethodGet {
			duration := time.Since(start)
			log.Printf("%s %s %s %s", req.RemoteAddr, req.Method, req.URL.Path, duration)
		}

	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
