package app

import (
	"net/http"

	"github.com/gjtiquia/google-calendar-analyzer/internal/web"
)

type Server struct {
	cfg Config
}

func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	webHandler := web.NewHandler()

	mux.HandleFunc("GET /", webHandler.Home)
	mux.HandleFunc("GET /healthz", webHandler.Healthz)

	return mux
}
