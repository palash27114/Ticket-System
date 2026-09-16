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
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.LoadConfig()
	if cfg.DeploymentMode == "serverless" {
		log.Fatal("DEPLOYMENT_MODE=serverless must be deployed through a serverless handler (for example, Vercel api/index.go)")
	}

	app, err := application.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Server listening on %s", addr)
	if err := app.Router.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
