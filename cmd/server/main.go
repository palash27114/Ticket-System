package main

import (
	"context"
	"fmt"
	"log"

	"ticket-system/internal/auth"
	"ticket-system/internal/config"
	"ticket-system/internal/database"
	"ticket-system/internal/health"
	"ticket-system/internal/server"
	"ticket-system/internal/ticket"
	"ticket-system/internal/user"
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

	log.Printf("Connecting to database at %s...", cfg.DatabaseURL)
	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(context.Background(), pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := user.NewPostgresRepository(pool)
	ticketRepo := ticket.NewPostgresRepository(pool)

	healthHandler := health.NewHandler(pool)
	authService := auth.NewService(userRepo, cfg)
	ticketService := ticket.NewService(ticketRepo)

	authHandler := auth.NewHandler(authService)
	ticketHandler := ticket.NewHandler(ticketService)

	r := server.SetupRouter(healthHandler, authHandler, ticketHandler, cfg.JWTSecret)

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
