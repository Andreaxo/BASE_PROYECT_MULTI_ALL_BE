package benefit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/benefit/application"
	"multicliente-backend/internal/features/benefit/domain"
	"multicliente-backend/internal/features/benefit/infrastructure"
	"multicliente-backend/internal/platform/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, requireRole gin.HandlerFunc) domain.BenefitRepository {
	repo := infrastructure.NewBenefitRepository(db)
	service := application.NewBenefitService(repo)
	handler := infrastructure.NewBenefitHandler(service, db)

	requireMembresia := middleware.RequireMembresiaActiva(db)
	requireSuscripcion := middleware.RequireSuscripcionActiva(db)

	Benefit := router.Group("/benefit")
	{
		// Authenticated user views (gated for affiliates)
		Benefit.GET("", requireMembresia, handler.GetAll)
		Benefit.GET("/my-company", handler.GetMyCompanyBenefits)
		Benefit.GET("/:id", requireMembresia, handler.GetBenefitByID)

		// Create/Edit/Delete operations (protected by permission, business subscription & empresa_id validation)
		Benefit.POST("", requireSuscripcion, middleware.RequirePermission(db, "/benefit", "CREATE"), handler.Create)
		Benefit.PUT("/:id", middleware.RequirePermission(db, "/benefit", "EDIT"), handler.Update)
		Benefit.DELETE("/:id", middleware.RequirePermission(db, "/benefit", "DELETE"), handler.Delete)
	}

	return repo
}
