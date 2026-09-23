package notificacion

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/notificacion/application"
	"multicliente-backend/internal/features/notificacion/domain"
	"multicliente-backend/internal/features/notificacion/infrastructure"
	"multicliente-backend/internal/platform/email"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes conecta la infraestructura de notificaciones, registra los endpoints protegidos por JWT,
// y retorna el NotificacionService para permitir su inyección en otros módulos (como membresías).
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, emailSvc email.EmailService, jwtSecret string) domain.NotificacionService {
	repo := infrastructure.NewNotificacionRepository(db)
	service := application.NewNotificacionService(repo, db, emailSvc)
	handler := infrastructure.NewNotificacionHandler(service)

	notifGroup := router.Group("/notificaciones")
	notifGroup.Use(middleware.JWTAuth(jwtSecret))
	{
		notifGroup.GET("", handler.GetNotificaciones)
		notifGroup.GET("/no-leidas/count", handler.GetCountNoLeidas)
		notifGroup.PUT("/:id/leido", handler.MarcarLeida)
		notifGroup.PUT("/leer-todas", handler.MarcarTodasLeidas)
	}

	return service
}
