package auth

import (
	"errors"
	"net/http"

	"ticket-system/internal/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body user.RegisterRequest true "User Registration Details"
// @Success 201 {object} user.RegisterResponse
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password and receive a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body user.LoginRequest true "User Login Credentials"
// @Success 200 {object} user.LoginResponse
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, user.LoginResponse{Token: token})
}
