package ticket

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ticket-system/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

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
