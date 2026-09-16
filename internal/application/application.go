package application

import (
	"context"

	"ticket-system/internal/auth"
	"ticket-system/internal/config"
	"ticket-system/internal/database"
	"ticket-system/internal/health"
	"ticket-system/internal/server"
	"ticket-system/internal/ticket"
	"ticket-system/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Application owns the shared dependencies used by the HTTP API.
type Application struct {
	Router *gin.Engine
	Pool   *pgxpool.Pool
}

// New connects the database, applies idempotent migrations, and wires the API.
func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	userRepo := user.NewPostgresRepository(pool)
	ticketRepo := ticket.NewPostgresRepository(pool)

	healthHandler := health.NewHandler(pool)
	authHandler := auth.NewHandler(auth.NewService(userRepo, cfg))
	ticketHandler := ticket.NewHandler(ticket.NewService(ticketRepo))

	return &Application{
		Router: server.SetupRouter(healthHandler, authHandler, ticketHandler, cfg.JWTSecret),
		Pool:   pool,
	}, nil
}

// Close releases the application's database resources.
func (a *Application) Close() {
	if a != nil && a.Pool != nil {
		a.Pool.Close()
	}
}
