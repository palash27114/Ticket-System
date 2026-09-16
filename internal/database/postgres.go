package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"ticket-system/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect initializes and returns a PostgreSQL connection pool.
func Connect(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")
	return pool, nil
}

// Migrate runs the embedded SQL migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Running database migrations...")
	if _, err := pool.Exec(ctx, migrations.InitSQL); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Database migrations applied successfully")
	return nil
}
