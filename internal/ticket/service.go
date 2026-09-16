package ticket

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidInput = errors.New("title and description cannot be empty")
	ErrTicketClosed = errors.New("ticket is closed and cannot be updated")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, req *CreateTicketRequest) (*Ticket, error) {
	title := strings.TrimSpace(req.Title)
	desc := strings.TrimSpace(req.Description)

	if title == "" || desc == "" {
		return nil, ErrInvalidInput
	}

	t := &Ticket{
		UserID:      userID,
		Title:       title,
		Description: desc,
		Status:      StatusOpen,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]*Ticket, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id int64, userID int64) (*Ticket, error) {
	return s.repo.GetByIDAndUserID(ctx, id, userID)
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, userID int64, newStatus string) (*Ticket, error) {
	// First fetch ticket to verify existence and check ownership
	currentTicket, err := s.repo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if currentTicket.Status == StatusClosed {
		return nil, ErrTicketClosed
	}

	// Validate state machine transition
	if !CanTransition(currentTicket.Status, newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	// Perform the update
	updatedTicket, err := s.repo.UpdateStatus(ctx, id, userID, newStatus)
	if err != nil {
		return nil, err
	}

	return updatedTicket, nil
}
