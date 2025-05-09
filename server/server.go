package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/config"
	"github.com/jeremyseow/unravel-be/server/handler"
	"github.com/jeremyseow/unravel-be/server/router"
	"github.com/jeremyseow/unravel-be/storage"
)

type Server struct {
	Cfg        *config.Config
	HTTPRouter *gin.Engine
}

func NewServer(cfg *config.Config, allStorages *storage.AllStorages) *Server {
	httpRouter := gin.Default()
	allHandlers := handler.NewAllHandlers(cfg, allStorages)
	router.SetupRoutes(httpRouter, allHandlers)

	return &Server{
		Cfg:        cfg,
		HTTPRouter: httpRouter,
	}
}

func (s *Server) StartServer() {
	go func() {
		if err := s.HTTPRouter.Run(fmt.Sprintf(":%d", s.Cfg.Server.Port)); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	fmt.Println("Server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	s.StopServer()
}

func (s *Server) StopServer() {
	fmt.Println("Server stopped")
}
