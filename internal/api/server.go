package api

import (
	"banner-rotation/internal/app"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	bandit app.BanditInterface
	server *http.Server
}

func NewServer(bandit app.BanditInterface) *Server {
	router := gin.Default()

	server := &Server{
		router: router,
		bandit: bandit,
	}

	server.setupRoutes()
	return server
}

func (s *Server) Start(address string) error {
	s.server = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
