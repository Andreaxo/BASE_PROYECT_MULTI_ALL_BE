package infrastructure

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/auth/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/i18n"
	"multicliente-backend/internal/platform/middleware"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	service domain.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	response, err := h.service.Login(&req)
	if err != nil {
		errStr := err.Error()
		if strings.HasPrefix(errStr, "cuenta_bloqueada:") {
			parts := strings.Split(errStr, ":")
			mins := "15"
			if len(parts) > 1 {
				mins = parts[1]
			}
			c.JSON(http.StatusLocked, gin.H{
				"error":     fmt.Sprintf("Tu cuenta ha sido bloqueada temporalmente por %s minutos debido a múltiples intentos fallidos de inicio de sesión.", mins),
				"is_locked": true,
				"minutes":   mins,
			})
			return
		}
		if strings.HasPrefix(errStr, "credenciales_invalidas:") {
			parts := strings.Split(errStr, ":")
			rem := "4"
			if len(parts) > 1 {
				rem = parts[1]
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":              fmt.Sprintf("Credenciales inválidas. Te quedan %s intentos antes de que tu cuenta sea bloqueada por 15 minutos.", rem),
				"remaining_attempts": rem,
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	// Set httpOnly cookie with dynamic role session duration
	c.SetCookie("token", response.Token, response.SessionDurationSeconds, "/", "", false, true)

	c.JSON(http.StatusOK, response)
}

// Logout handles POST /api/auth/logout.
// Invalidates the JWT token in the blacklist and destroys the httpOnly session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenString := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			tokenString = parts[1]
		}
	}
	if tokenString == "" {
		if cookieToken, err := c.Cookie("token"); err == nil {
			tokenString = cookieToken
		}
	}

	if tokenString != "" {
		middleware.GlobalTokenBlacklist.Revoke(tokenString, time.Now().Add(24*time.Hour))
	}

	// Destroy session cookie in browser
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada correctamente"})
}

// Register handles POST /api/auth/register (public endpoint).
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	response, err := h.service.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	// Set httpOnly cookie for auto-login after registration
	c.SetCookie("token", response.Token, response.SessionDurationSeconds, "/", "", false, true)

	c.JSON(http.StatusCreated, response)
}

// GetMyCodeRefer handles GET /api/auth/mi-codigo-referido (authenticated).
// Returns only the authenticated user's own referral code.
func (h *AuthHandler) GetMyCodeRefer(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
		return
	}

	var userID uint
	if f, ok := userIDVal.(float64); ok {
		userID = uint(f)
	} else if u, ok := userIDVal.(uint); ok {
		userID = u
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, errors.New("invalid user ID format"))})
		return
	}

	code, err := h.service.GetMyCodeRefer(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code_refer": code})
}

// GetProfile handles GET /api/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
		return
	}

	var userID uint
	if f, ok := userIDVal.(float64); ok {
		userID = uint(f)
	} else if u, ok := userIDVal.(uint); ok {
		userID = u
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, errors.New("invalid user ID format in session"))})
		return
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, userDomain.ToUserResponse(user))
}

// ChangePassword handles PUT /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
		return
	}

	var userID uint
	if f, ok := userIDVal.(float64); ok {
		userID = uint(f)
	} else if u, ok := userIDVal.(uint); ok {
		userID = u
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, errors.New("invalid user ID format"))})
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	if err := h.service.ChangePassword(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contraseña cambiada correctamente"})
}

// UpdateProfile handles PUT /api/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
		return
	}

	var userID uint
	if f, ok := userIDVal.(float64); ok {
		userID = uint(f)
	} else if u, ok := userIDVal.(uint); ok {
		userID = u
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, errors.New("invalid user ID format"))})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	user, err := h.service.UpdateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, userDomain.ToUserResponse(user))
}

// ValidarCodigoEmpresa handles GET /api/auth/validar-codigo-empresa/:codigo (public)
func (h *AuthHandler) ValidarCodigoEmpresa(c *gin.Context) {
	codigo := c.Param("codigo")
	if codigo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código de empresa requerido"})
		return
	}

	company, err := h.service.ValidarCodigoEmpresa(codigo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             company.ID,
		"name":           company.Name,
		"codigo_empresa": company.CodigoEmpresa,
	})
}

// OlvidePassword handles POST /api/auth/olvide-password (public)
func (h *AuthHandler) OlvidePassword(c *gin.Context) {
	var req domain.OlvidePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	if err := h.service.OlvidePassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Se ha enviado un correo electrónico con las instrucciones para restablecer tu contraseña.",
	})
}

// VerificarCodigo handles POST /api/auth/verificar-codigo (public)
func (h *AuthHandler) VerificarCodigo(c *gin.Context) {
	var req domain.VerificarCodigoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	response, err := h.service.VerificarCodigo(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResetPassword handles POST /api/auth/reset-password (public)
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	if err := h.service.ResetPassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tu contraseña ha sido actualizada con éxito. Ya puedes iniciar sesión con tus nuevas credenciales.",
	})
}

// RegistroAsistido handles POST /api/auth/registro-asistido (protected, operador or superadmin)
func (h *AuthHandler) RegistroAsistido(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
		return
	}

	var operadorID uint
	if f, ok := userIDVal.(float64); ok {
		operadorID = uint(f)
	} else if u, ok := userIDVal.(uint); ok {
		operadorID = u
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, errors.New("invalid user ID format"))})
		return
	}

	var req domain.RegistroAsistidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	response, err := h.service.RegistroAsistido(&req, operadorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusCreated, response)
}


