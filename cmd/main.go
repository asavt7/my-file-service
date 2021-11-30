package main

import (
	"github.com/asavt7/my-file-service/internal/app"
	"github.com/asavt7/my-file-service/internal/config"
	log "github.com/sirupsen/logrus"
)

// @title github.com/asavt7/my-file-service
// @version 1.0
// @description This is a simple image upload service

// @contact.name https://github.com/asavt7
// @contact.url https://github.com/asavt7

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("ERROR init configs %+v", err)
	}

	app.NewApp(cfg).Run()

}
