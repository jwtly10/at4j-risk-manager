package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jwtly10/at4j-risk-manager/internal/api/handlers"
	"github.com/jwtly10/at4j-risk-manager/internal/api/middleware"
	"github.com/jwtly10/at4j-risk-manager/internal/config"
	"github.com/jwtly10/at4j-risk-manager/internal/db"
	"github.com/jwtly10/at4j-risk-manager/pkg/logger"
)

type Server struct {
	config     *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config, dbClient *db.Client) *Server {
	equityHandler := handlers.NewEquityHandler(dbClient)

	mux := http.NewServeMux()

	auth := middleware.APIKeyAuth(cfg.ApiKey)
	mux.HandleFunc("/api/v1/equity/latest", auth(equityHandler.GetLatestEquity))
	mux.HandleFunc("/health", handlers.HealthCheck)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	return &Server{
		config:     cfg,
		httpServer: server,
	}
}

func (s *Server) Start() error {
	logger.Infof("Starting HTTP server on port %s", s.config.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
