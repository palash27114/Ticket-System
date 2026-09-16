package server

import (
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ticket-system/docs"
	"ticket-system/internal/auth"
	"ticket-system/internal/health"
	"ticket-system/internal/middleware"
	"ticket-system/internal/ticket"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes Gin router with all application routes and middleware.
func SetupRouter(
	healthHandler *health.Handler,
	authHandler *auth.Handler,
	ticketHandler *ticket.Handler,
	jwtSecret string,
) *gin.Engine {
	r := gin.Default()

	// Swagger UI — redirect /swagger to /swagger/index.html for convenience
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint (no authentication required)
	r.GET("/health", healthHandler.Check)

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
