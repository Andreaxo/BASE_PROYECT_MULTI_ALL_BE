package auth

import (
	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/auth/application"
	"multicliente-backend/internal/features/auth/infrastructure"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/middleware"
)

// RegisterRoutes wires up the auth feature and registers its routes.
// It receives the user repository and referido service as cross-feature dependencies.
func RegisterRoutes(
	router *gin.RouterGroup,
	userRepo userDomain.UserRepository,
	referidoSvc referidoDomain.ReferidoService,
	jwtSecret string,
	jwtExpHours string,
) {
	service := application.NewAuthService(userRepo, referidoSvc, jwtSecret, jwtExpHours)
	handler := infrastructure.NewAuthHandler(service)

	authGroup := router.Group("/auth")
	{
		// Public endpoints (no JWT required)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/register", handler.Register)

		// Protected endpoints (JWT required)
		authGroup.GET("/profile", middleware.JWTAuth(jwtSecret), handler.GetProfile)
		authGroup.PUT("/profile", middleware.JWTAuth(jwtSecret), handler.UpdateProfile)
		authGroup.GET("/mi-codigo-referido", middleware.JWTAuth(jwtSecret), handler.GetMyCodeRefer)
		authGroup.PUT("/change-password", middleware.JWTAuth(jwtSecret), handler.ChangePassword)
	}
}
