package infrastructure

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/auth/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/i18n"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	// Set httpOnly cookie with dynamic role session duration
	c.SetCookie("token", response.Token, response.SessionDurationSeconds, "/", "", false, true)

	c.JSON(http.StatusOK, response)
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

