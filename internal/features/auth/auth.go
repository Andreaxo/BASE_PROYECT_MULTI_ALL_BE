package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/auth/application"
	"multicliente-backend/internal/features/auth/infrastructure"
	companyDomain "multicliente-backend/internal/features/company/domain"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/email"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the auth feature and registers its routes.
// It receives the user repository, referido service, company repository, db, and email service as cross-feature dependencies.
func RegisterRoutes(
	router *gin.RouterGroup,
	userRepo userDomain.UserRepository,
	referidoSvc referidoDomain.ReferidoService,
	companyRepo companyDomain.CompanyRepository,
	db *gorm.DB,
	emailSvc email.EmailService,
	jwtSecret string,
	jwtExpHours string,
) {
	service := application.NewAuthService(userRepo, referidoSvc, companyRepo, db, emailSvc, jwtSecret, jwtExpHours)
	handler := infrastructure.NewAuthHandler(service)

	authGroup := router.Group("/auth")
	{
		// Public endpoints (no JWT required)
		authGroup.POST("/login", middleware.LoginRateLimit(), handler.Login)
		authGroup.POST("/register", handler.Register)
		authGroup.GET("/validar-codigo-empresa/:codigo", handler.ValidarCodigoEmpresa)
		authGroup.POST("/olvide-password", handler.OlvidePassword)
		authGroup.POST("/verificar-codigo", handler.VerificarCodigo)
		authGroup.POST("/reset-password", handler.ResetPassword)
		authGroup.POST("/logout", handler.Logout)

		// Protected endpoints (JWT required)
		authGroup.GET("/profile", middleware.JWTAuth(jwtSecret), handler.GetProfile)
		authGroup.PUT("/profile", middleware.JWTAuth(jwtSecret), handler.UpdateProfile)
		authGroup.GET("/mi-codigo-referido", middleware.JWTAuth(jwtSecret), handler.GetMyCodeRefer)
		authGroup.PUT("/change-password", middleware.JWTAuth(jwtSecret), handler.ChangePassword)

		// Operator-assisted registration (requires operador or superadmin)
		authGroup.POST(
			"/registro-asistido",
			middleware.JWTAuth(jwtSecret),
			middleware.RequireRole("operador", "superadmin"),
			handler.RegistroAsistido,
		)
	}
}

