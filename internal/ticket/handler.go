package ticket

import (
	"errors"
	"net/http"
	"strconv"

	"ticket-system/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary Create a new ticket
// @Description Create a new ticket for the authenticated user with initial status 'open'
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTicketRequest true "Ticket Details"
// @Success 201 {object} Ticket
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tickets [post]
func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// List godoc
// @Summary List user's tickets
// @Description List all tickets belonging to the authenticated user ordered by created_at DESC
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Ticket
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tickets [get]
func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tickets, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// GetByID godoc
// @Summary Get ticket by ID
// @Description Get a specific ticket belonging to the authenticated user by ticket ID
// @Tags Tickets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} Ticket
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tickets/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	ticket, err := h.service.GetByID(c.Request.Context(), ticketID, userID)
	if err != nil {
		if errors.Is(err, ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve ticket"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// UpdateStatus godoc
// @Summary Update ticket status
// @Description Update the status of a ticket belonging to the authenticated user following the state machine: open -> in_progress -> closed
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body UpdateStatusRequest true "New Ticket Status"
// @Success 200 {object} Ticket
// @Failure 400 {object} map[string]string "Invalid status transition"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Ticket not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /tickets/{id}/status [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.UpdateStatus(c.Request.Context(), ticketID, userID, req.Status)
	if err != nil {
		if errors.Is(err, ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if errors.Is(err, ErrInvalidStatusTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update ticket status"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}
