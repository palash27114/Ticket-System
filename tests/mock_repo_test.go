package tests

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"ticket-system/internal/auth"
	"ticket-system/internal/config"
	"ticket-system/internal/server"
	"ticket-system/internal/ticket"
	"ticket-system/internal/user"
)

type InMemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[int64]*user.User
	emails map[string]*user.User
	seq    int64
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users:  make(map[int64]*user.User),
		emails: make(map[string]*user.User),
	}
}

func (m *InMemoryUserRepo) Create(ctx context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	email := strings.ToLower(strings.TrimSpace(u.Email))
	if _, exists := m.emails[email]; exists {
		return user.ErrEmailAlreadyExists
	}

	m.seq++
	u.ID = m.seq
	u.CreatedAt = time.Now()

	userCopy := *u
	m.users[u.ID] = &userCopy
	m.emails[email] = &userCopy
	return nil
}

func (m *InMemoryUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	u, exists := m.emails[normalizedEmail]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	userCopy := *u
	return &userCopy, nil
}

func (m *InMemoryUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.users[id]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	userCopy := *u
	return &userCopy, nil
}

type InMemoryTicketRepo struct {
	mu      sync.RWMutex
	tickets map[int64]*ticket.Ticket
	seq     int64
}

func NewInMemoryTicketRepo() *InMemoryTicketRepo {
	return &InMemoryTicketRepo{
		tickets: make(map[int64]*ticket.Ticket),
	}
}

func (m *InMemoryTicketRepo) Create(ctx context.Context, t *ticket.Ticket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seq++
	t.ID = m.seq
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	ticketCopy := *t
	m.tickets[t.ID] = &ticketCopy
	return nil
}

func (m *InMemoryTicketRepo) ListByUserID(ctx context.Context, userID int64) ([]*ticket.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*ticket.Ticket, 0)
	for _, t := range m.tickets {
		if t.UserID == userID {
			ticketCopy := *t
			results = append(results, &ticketCopy)
		}
	}

	// Reverse order to match ORDER BY created_at DESC
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

func (m *InMemoryTicketRepo) GetByIDAndUserID(ctx context.Context, id int64, userID int64) (*ticket.Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.tickets[id]
	if !exists || t.UserID != userID {
		return nil, ticket.ErrTicketNotFound
	}

	ticketCopy := *t
	return &ticketCopy, nil
}

func (m *InMemoryTicketRepo) UpdateStatus(ctx context.Context, id int64, userID int64, status string) (*ticket.Ticket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tickets[id]
	if !exists || t.UserID != userID {
		return nil, ticket.ErrTicketNotFound
	}

	t.Status = status
	t.UpdatedAt = time.Now()

	ticketCopy := *t
	return &ticketCopy, nil
}

// setupTestServer initializes a test router with in-memory repositories
func setupTestServer() (*gin.Engine, *config.Config) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Port:               "8080",
		DatabaseURL:        "mock",
		JWTSecret:          "test-secret-key-12345",
		JWTExpirationHours: 24,
	}

	userRepo := NewInMemoryUserRepo()
	ticketRepo := NewInMemoryTicketRepo()

	authService := auth.NewService(userRepo, cfg)
	ticketService := ticket.NewService(ticketRepo)

	authHandler := auth.NewHandler(authService)
	ticketHandler := ticket.NewHandler(ticketService)

	r := server.SetupRouter(authHandler, ticketHandler, cfg.JWTSecret)
	return r, cfg
}
