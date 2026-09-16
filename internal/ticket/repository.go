package ticket

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, t *Ticket) error
	ListByUserID(ctx context.Context, userID int64) ([]*Ticket, error)
	GetByIDAndUserID(ctx context.Context, id int64, userID int64) (*Ticket, error)
	UpdateStatus(ctx context.Context, id int64, userID int64, status string) (*Ticket, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, t *Ticket) error {
	query := `
		INSERT INTO tickets (user_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query, t.UserID, t.Title, t.Description, t.Status).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *PostgresRepository) ListByUserID(ctx context.Context, userID int64) ([]*Ticket, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tickets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]*Ticket, 0)
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *PostgresRepository) GetByIDAndUserID(ctx context.Context, id int64, userID int64) (*Ticket, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tickets
		WHERE id = $1 AND user_id = $2
	`
	var t Ticket
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id int64, userID int64, status string) (*Ticket, error) {
	query := `
		UPDATE tickets
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, title, description, status, created_at, updated_at
	`
	var t Ticket
	err := r.pool.QueryRow(ctx, query, status, id, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	return &t, nil
}
