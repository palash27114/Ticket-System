package ticket

import (
	"errors"
	"time"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

var (
	ErrTicketNotFound          = errors.New("ticket not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type Ticket struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTicketRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// IsValidStatus checks if the given status is a recognized ticket status.
func IsValidStatus(status string) bool {
	return status == StatusOpen || status == StatusInProgress || status == StatusClosed
}

// CanTransition validates the status transition state machine:
// open -> in_progress -> closed
// All other transitions (including reopening or same-state transitions) are invalid.
func CanTransition(currentStatus, newStatus string) bool {
	if currentStatus == StatusOpen && newStatus == StatusInProgress {
		return true
	}
	if currentStatus == StatusInProgress && newStatus == StatusClosed {
		return true
	}
	return false
}
