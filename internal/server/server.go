package server

import (
	"fmt"
	"github.com/asavt7/my-file-service/internal/config"
	"github.com/asavt7/my-file-service/internal/health"
	"github.com/asavt7/my-file-service/internal/store"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const MaxUploadSize = 1024 * 1024 * 10 // 10MB

type APIServer struct {
	store  store.Store
	config config.ServerConfig

	health health.Healthchecker
}

func NewAPIServer(s store.Store, config config.ServerConfig, h health.Healthchecker) *APIServer {
	return &APIServer{store: s, config: config, health: h}
}

func (s *APIServer) Run() error {
	mainHandler := s.initHandlers()
	log.Infof("Starting server at :%d", s.config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), mainHandler)
	return err
}
