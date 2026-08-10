package benefit_redemption

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/benefit_redemption/application"
	"multicliente-backend/internal/features/benefit_redemption/domain"
	"multicliente-backend/internal/features/benefit_redemption/infrastructure"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the benefit_redemption feature: repository → service → handler → routes.
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) domain.RedemptionService {
	repo := infrastructure.NewRedemptionRepository(db)
	service := application.NewRedemptionService(repo)
	handler := infrastructure.NewRedemptionHandler(service)

	// Employee endpoints — any authenticated user
	benefits := router.Group("/benefits")
	{
		benefits.GET("/mis-redenciones", handler.GetMisRedenciones)
		benefits.POST("/:id/redeem", handler.RedeemBenefit)
	}

	// Business validator endpoints — requires business_validator role
	redemptions := router.Group("/redemptions")
	redemptions.Use(middleware.RequireRole("business_validator"))
	{
		redemptions.POST("/validate", handler.ValidateCode)
		redemptions.GET("", handler.GetRedemptionsByEmpresa)
	}

	// Admin endpoint — requires permission on /redemptions menu
	admin := router.Group("/admin")
	{
		admin.GET("/redemptions", middleware.RequirePermission(db, "/redemptions", "VIEW"), handler.GetAllRedemptions)
	}

	return service
}
