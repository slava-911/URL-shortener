package main

import (
	"log"

	"github.com/slava-911/URL-shortener/internal/app"
	"github.com/slava-911/URL-shortener/internal/config"
	"github.com/slava-911/URL-shortener/pkg/logging"
)

func main() {
	log.Print("config initializing")
	cfg := config.GetConfig()

	log.Print("logger initializing")
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)

	a, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("Running Application")
	a.Run()
}
