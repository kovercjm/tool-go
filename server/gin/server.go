package gin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kovercjm/tool-go/logger"
	"github.com/kovercjm/tool-go/server"
)

const (
	CtxToken = "token"
)

type Server struct {
	GinEngine  *gin.Engine
	HTTPServer *http.Server

	config *server.APIConfig
	logger logger.Logger
}

func NewServer(config *server.Config, logger logger.Logger) *Server {
	if config == nil || logger == nil {
		panic("Missing critical arguments to init a server")
	}
	s := Server{
		GinEngine: gin.New(),
		config:    &server.APIConfig{Port: config.APIConfig.Port},
		logger:    logger.NoCaller(),
	}
	return &s
}

func (s *Server) WithDefaultMiddlewares() *Server {
	s.GinEngine.Use(
		APILogging(s.logger),
		ErrorFormatter(),
		PanicRecovery(s.logger),
	)
	return s
}

func (s *Server) Start(_ context.Context) error {
	address := fmt.Sprintf(":%d", s.config.Port)
	s.HTTPServer = &http.Server{
		Addr:    address,
		Handler: s.GinEngine,
	}
	go func() {
		s.logger.Info("gin api server starting", "listening", address)
		if err := s.HTTPServer.ListenAndServe(); err != nil {
			s.logger.Error("gin api server failed to serve", "error", err)
		}
	}()
	return nil
}

func (s *Server) Stop(_ context.Context) error {
	s.logger.Info("gin api server is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		s.logger.Error("gin api server shutdown failed", "error", err)
	}

	s.logger.Info("gin api server stopped gracefully")
	return nil
}
