package main

import (
	"log"

	"project_sem/internal/server"
	"project_sem/platform/config"
)

func main() {
	settings, err := config.Load("")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	app, err := server.New(settings)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	app.Run()
}