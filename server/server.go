package server

import (
	"fmt"
	"net/http"
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
	go http.ListenAndServe(fmt.Sprintf("%s:%d", s.Cfg.Server.Host, s.Cfg.Server.Port), s.HTTPRouter)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	s.StopServer()
}

func (s *Server) StopServer() {
	fmt.Println("Server stopped")
}
