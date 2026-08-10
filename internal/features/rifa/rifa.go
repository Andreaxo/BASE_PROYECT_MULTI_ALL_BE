package rifa

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	referidoInfra "multicliente-backend/internal/features/referido/infrastructure"
	"multicliente-backend/internal/features/rifa/application"
	"multicliente-backend/internal/features/rifa/domain"
	"multicliente-backend/internal/features/rifa/infrastructure"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the rifa feature: repositories → service → handler → routes.
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) domain.RifaService {
	rifaRepo := infrastructure.NewRifaRepository(db)
	participacionRepo := infrastructure.NewParticipacionRepository(db)
	ganadorRepo := infrastructure.NewGanadorRepository(db)
	referidoRepo := referidoInfra.NewReferidoRepository(db)

	service := application.NewRifaService(db, rifaRepo, participacionRepo, ganadorRepo, referidoRepo)
	handler := infrastructure.NewRifaHandler(service)

	rifas := router.Group("/rifas")
	{
		// User endpoints
		rifas.GET("/activa", handler.GetActiva)
		rifas.GET("/mis-participaciones", handler.GetMisParticipaciones)
		rifas.GET("/historial", handler.GetHistorial)
		rifas.POST("/:id/participar", handler.Participar)

		// Admin endpoints
		rifas.GET("", middleware.RequirePermission(db, "/rifas", "VIEW"), handler.GetAll)
		rifas.GET("/:id", middleware.RequirePermission(db, "/rifas", "VIEW"), handler.GetByID)
		rifas.POST("", middleware.RequirePermission(db, "/rifas", "CREATE"), handler.Create)
		rifas.PUT("/:id", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Update)
		rifas.PUT("/:id/activar", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Activar)
		rifas.PUT("/:id/cerrar", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Cerrar)
		rifas.PUT("/:id/sortear", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Sortear)
		rifas.POST("/:id/sortear", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Sortear)
		rifas.PUT("/:id/cancelar", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.Cancelar)

		rifas.GET("/:id/participaciones", middleware.RequirePermission(db, "/rifas", "VIEW"), handler.GetParticipaciones)
		rifas.POST("/:id/participaciones", middleware.RequirePermission(db, "/rifas", "CREATE"), handler.AgregarParticipacion)

		rifas.GET("/:id/ganadores", middleware.RequirePermission(db, "/rifas", "VIEW"), handler.GetGanadores)
		rifas.POST("/:id/ganadores", middleware.RequirePermission(db, "/rifas", "CREATE"), handler.RegistrarGanador)
		rifas.PUT("/ganadores/:id/entrega", middleware.RequirePermission(db, "/rifas", "EDIT"), handler.ActualizarEntrega)
	}

	return service
}
