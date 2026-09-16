package main

import (
	"context"
	"fmt"
	"log"

	"ticket-system/internal/auth"
	"ticket-system/internal/config"
	"ticket-system/internal/database"
	"ticket-system/internal/server"
	"ticket-system/internal/ticket"
	"ticket-system/internal/user"
)

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

	authService := auth.NewService(userRepo, cfg)
	ticketService := ticket.NewService(ticketRepo)

	authHandler := auth.NewHandler(authService)
	ticketHandler := ticket.NewHandler(ticketService)

	r := server.SetupRouter(authHandler, ticketHandler, cfg.JWTSecret)

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
