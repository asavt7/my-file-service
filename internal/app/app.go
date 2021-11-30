package app

import (
	"github.com/asavt7/my-file-service/internal/config"
	"github.com/asavt7/my-file-service/internal/server"
	"github.com/asavt7/my-file-service/internal/store"
	log "github.com/sirupsen/logrus"
)

type App struct {
	cfg *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run() {
	storage, err := store.NewAwsStore(&a.cfg.S3Config)
	if err != nil {
		log.Fatalf("Cannot init AWS store client %+v\n", err)
	}

	APIServer := server.NewAPIServer(storage, a.cfg.ServerConfig, storage)
	err = APIServer.Run()
	if err != nil {
		log.Fatalf("Failed to start server : %+v\n", err)
	}
}
