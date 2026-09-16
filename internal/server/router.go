package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ticket-system/internal/auth"
	"ticket-system/internal/middleware"
	"ticket-system/internal/ticket"
)

// SetupRouter initializes Gin router with all application routes and middleware.
func SetupRouter(authHandler *auth.Handler, ticketHandler *ticket.Handler, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Health check endpoint (No authentication)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Public authentication routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected ticket routes
	ticketsGroup := r.Group("/tickets")
	ticketsGroup.Use(middleware.AuthMiddleware(jwtSecret))
	{
		ticketsGroup.POST("", ticketHandler.Create)
		ticketsGroup.GET("", ticketHandler.List)
		ticketsGroup.GET("/:id", ticketHandler.GetByID)
		ticketsGroup.PATCH("/:id/status", ticketHandler.UpdateStatus)
	}

	return r
}
