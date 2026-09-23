package server

import (
	"multicliente-backend/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures a new Gin router with global middlewares.
func NewRouter() *gin.Engine {
	router := gin.Default()

	// Apply global middlewares
	router.Use(middleware.CORS())
	router.Use(middleware.Translation())
	// Security: block SQL injection patterns on every request
	router.Use(middleware.SQLInjectionGuard())

	// Serve uploaded profile and company images
	router.Static("/uploads", "./uploads")

	return router
}
