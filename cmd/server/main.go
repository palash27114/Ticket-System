package main

import (
	"context"
	"fmt"
	"log"

	"ticket-system/internal/application"
	"ticket-system/internal/config"
)

// @title Ticket System API
// @version 1.0
// @description REST API for Backend Intern Ticket System with JWT authentication and ticket status state machine.
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.LoadConfig()
	if cfg.DeploymentMode != "local" && cfg.DeploymentMode != "vercel" {
		log.Fatalf("unsupported DEPLOYMENT_MODE %q; use local or vercel", cfg.DeploymentMode)
	}

	app, err := application.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Server listening on %s (mode: %s)", addr, cfg.DeploymentMode)
	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
