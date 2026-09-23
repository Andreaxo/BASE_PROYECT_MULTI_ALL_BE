package membresia

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	referidoInfra "multicliente-backend/internal/features/referido/infrastructure"
	userDomain "multicliente-backend/internal/features/user/domain"

	"multicliente-backend/internal/features/membresia/application"
	"multicliente-backend/internal/features/membresia/domain"
	"multicliente-backend/internal/features/membresia/infrastructure"
	notificacionDomain "multicliente-backend/internal/features/notificacion/domain"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the membresia feature: repositories → service → handlers → routes.
// Returns the MembresiaService for cross-feature use and the WebhookHandler for public routes.
func RegisterRoutes(
	router *gin.RouterGroup,
	publicRouter *gin.RouterGroup,
	db *gorm.DB,
	userRepo userDomain.UserRepository,
	notifSvc notificacionDomain.NotificacionService,
	wompiPublicKey string,
	wompiPrivateKey string,
	wompiEventsSecret string,
	wompiIntegritySecret string,
	wompiSandbox bool,
) domain.MembresiaService {
	membRepo := infrastructure.NewMembresiaRepository(db)
	pagoRepo := infrastructure.NewPagoMembresiaRepository(db)
	referidoRepo := referidoInfra.NewReferidoRepository(db)
	wompiClient := infrastructure.NewWompiClient(wompiPublicKey, wompiPrivateKey, wompiEventsSecret, wompiIntegritySecret, wompiSandbox)

	service := application.NewMembresiaService(db, membRepo, pagoRepo, referidoRepo, userRepo, wompiClient, notifSvc)
	handler := infrastructure.NewMembresiaHandler(service)
	webhookHandler := infrastructure.NewWebhookHandler(service)

	// Protected endpoints (JWT required)
	memb := router.Group("/membresia")
	{
		memb.POST("/iniciar-pago", handler.IniciarPago)
		memb.POST("/simular-pago", handler.SimularPago)
		memb.GET("/mi-membresia", handler.GetMiMembresia)
		memb.PUT("/cancelar-renovacion", handler.CancelarRenovacion)

		// Admin endpoint
		memb.GET("/admin/todas", middleware.RequirePermission(db, "/membresias", "VIEW"), handler.GetAll)
	}

	// Public webhook endpoint (no JWT)
	webhooks := publicRouter.Group("/webhooks")
	{
		webhooks.POST("/wompi", webhookHandler.HandleWompiWebhook)
	}

	return service
}
