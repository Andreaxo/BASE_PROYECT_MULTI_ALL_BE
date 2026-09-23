package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/platform/i18n"
)

// RequireMembresiaActiva returns a Gin middleware that checks if the authenticated user
// (with role 'user' or 'user_member') has an active membership.
// Superadmin is exempt. business_validator is skipped (handled by RequireSuscripcionActiva).
func RequireMembresiaActiva(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: role not found"))})
			c.Abort()
			return
		}
		role := strings.ToLower(strings.TrimSpace(roleVal.(string)))

		// Superadmin, admin and operador always pass unconditionally
		if role == "superadmin" || role == "super_admin" || role == "admin" || role == "operador" {
			c.Next()
			return
		}

		// business_validator is not gated here — handled by RequireSuscripcionActiva
		if role == "business_validator" || role == "negocio" {
			c.Next()
			return
		}

		// For user / user_member: check membership
		userID := extractUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
			c.Abort()
			return
		}

		var estado string
		err := db.Table("administrative.membresia").
			Select("estado").
			Where("usuario_id = ?", userID).
			Scan(&estado).Error

		if err != nil || estado != "activa" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": i18n.TranslateError(c, errors.New("membership_not_active")),
				"code":  "MEMBERSHIP_NOT_ACTIVE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuscripcionActiva returns a Gin middleware that checks if the authenticated
// business_validator's company has an active subscription (prueba or activa).
// Non-business_validator roles pass through (handled by other middleware).
func RequireSuscripcionActiva(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.Next()
			return
		}
		role := strings.ToLower(strings.TrimSpace(roleVal.(string)))

		// Only applies to business_validator / negocio
		if role != "business_validator" && role != "negocio" {
			c.Next()
			return
		}

		userID := extractUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("user not authenticated"))})
			c.Abort()
			return
		}

		// 1. Check if user's membership is active in administrative.membresia
		var membEstado string
		_ = db.Table("administrative.membresia").
			Select("estado").
			Where("usuario_id = ?", userID).
			Limit(1).
			Scan(&membEstado).Error

		// 2. Find the company's subscription state via user_companies OR users.empresa_id
		var suscripcionEstado string
		err := db.Table("administrative.companies c").
			Select("c.suscripcion_estado").
			Joins("LEFT JOIN administrative.user_companies uc ON uc.company_id = c.id").
			Joins("LEFT JOIN administrative.users u ON u.empresa_id = c.id").
			Where("uc.user_id = ? OR u.id = ?", userID, userID).
			Limit(1).
			Scan(&suscripcionEstado).Error

		// If user's membership is active, treat company as active and sync status
		if membEstado == "activa" {
			suscripcionEstado = "activa"
			_ = db.Exec(`
				UPDATE administrative.companies SET suscripcion_estado = 'activa'
				WHERE id IN (
					SELECT company_id FROM administrative.user_companies WHERE user_id = ?
					UNION
					SELECT empresa_id FROM administrative.users WHERE id = ? AND empresa_id IS NOT NULL
				)
			`, userID, userID)
		}

		if (err != nil && membEstado != "activa") || (suscripcionEstado != "prueba" && suscripcionEstado != "activa") {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("business_subscription_expired"))})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractUserID gets the user ID from Gin context, handling float64/string types.
func extractUserID(c *gin.Context) uint {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	if f, ok := userIDVal.(float64); ok {
		return uint(f)
	}

	if s, ok := userIDVal.(string); ok {
		parsed, err := strconv.ParseUint(s, 10, 32)
		if err == nil {
			return uint(parsed)
		}
	}

	return 0
}
