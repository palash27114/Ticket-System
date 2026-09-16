package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"ticket-system/internal/application"
	"ticket-system/internal/config"
)

var (
	initializeOnce sync.Once
	app            *application.Application
	initializeErr  error
)

func initialize() {
	cfg := config.LoadConfig()
	if cfg.DeploymentMode != "serverless" {
		initializeErr = fmt.Errorf("DEPLOYMENT_MODE must be set to serverless")
		return
	}

	app, initializeErr = application.New(context.Background(), cfg)
}

// Handler is the Vercel serverless entry point. Initialization is reused by warm invocations.
func Handler(w http.ResponseWriter, r *http.Request) {
	initializeOnce.Do(initialize)
	if initializeErr != nil {
		log.Printf("application initialization failed: %v", initializeErr)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	app.Router.ServeHTTP(w, r)
}
