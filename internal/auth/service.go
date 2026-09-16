package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"ticket-system/internal/config"
	"ticket-system/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	userRepo user.Repository
	cfg      *config.Config
}

func NewService(userRepo user.Repository, cfg *config.Config) *Service {
	return &Service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *Service) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &user.RegisterResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}, nil
}

func (s *Service) Login(ctx context.Context, req *user.LoginRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := GenerateToken(u.ID, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return "", err
	}

	return token, nil
}

