package referido

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/referido/application"
	"multicliente-backend/internal/features/referido/domain"
	"multicliente-backend/internal/features/referido/infrastructure"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the referido feature: repository → service → handler → routes.
// Returns the ReferidoService so other features (e.g., auth/register) can reuse it.
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) domain.ReferidoService {
	repo := infrastructure.NewReferidoRepository(db)
	service := application.NewReferidoService(repo)
	handler := infrastructure.NewReferidoHandler(service)

	requireMembresia := middleware.RequireMembresiaActiva(db)

	referidos := router.Group("/referidos")
	{
		// Authenticated user endpoints (gated by active membership for affiliates)
		referidos.POST("", requireMembresia, handler.Create)
		referidos.GET("/mis-referidos", requireMembresia, handler.GetMisReferidos)

		// Admin-only endpoints
		referidos.GET("", middleware.RequirePermission(db, "/referidos", "VIEW"), handler.GetAll)
		referidos.GET("/:id", middleware.RequirePermission(db, "/referidos", "VIEW"), handler.GetByID)
		referidos.PUT("/:id/afiliar", middleware.RequirePermission(db, "/referidos", "EDIT"), handler.Afiliar)
		referidos.PUT("/:id/recompensa", middleware.RequirePermission(db, "/referidos", "EDIT"), handler.Recompensa)
		referidos.DELETE("/:id", middleware.RequirePermission(db, "/referidos", "DELETE"), handler.Delete)
	}

	return service
}
