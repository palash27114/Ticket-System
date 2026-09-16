package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceStatus holds the health status of a single service.
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthResponse is the full response body for GET /health.
type HealthResponse struct {
	Status   string                   `json:"status"`
	Services map[string]ServiceStatus `json:"services"`
}

// Handler handles the health check endpoint.
type Handler struct {
	db *pgxpool.Pool
}

// NewHandler creates a new health Handler.
func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// Check godoc
// @Summary Health check
// @Description Returns the health status of the API server and all dependent services (PostgreSQL database).
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse "All services healthy"
// @Success 207 {object} HealthResponse "One or more services degraded"
// @Router /health [get]
func (h *Handler) Check(c *gin.Context) {
	services := make(map[string]ServiceStatus)
	overallOK := true

	// Check API service (always available if this handler runs)
	services["api"] = ServiceStatus{Status: "ok"}

	// Check PostgreSQL database
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			services["database"] = ServiceStatus{
				Status:  "error",
				Message: "database unreachable",
			}
			overallOK = false
		} else {
			services["database"] = ServiceStatus{Status: "ok"}
		}
	} else {
		services["database"] = ServiceStatus{
			Status:  "unknown",
			Message: "no database connection configured",
		}
	}

	resp := HealthResponse{
		Status:   "ok",
		Services: services,
	}

	httpStatus := http.StatusOK
	if !overallOK {
		resp.Status = "degraded"
		httpStatus = http.StatusMultiStatus // 207 - partial success
	}

	c.JSON(httpStatus, resp)
}
