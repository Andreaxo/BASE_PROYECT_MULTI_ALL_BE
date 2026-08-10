package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"multicliente-backend/internal/platform/i18n"
)

// RequirePermission returns a Gin middleware that checks if the authenticated user's role has permission for a specific menu and option.
func RequirePermission(db *gorm.DB, menuRoute string, optionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get role from context
		roleCodeVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: role not found in context"))})
			c.Abort()
			return
		}
		roleCode := roleCodeVal.(string)

		// Superadmin always has access. Business validator has access to benefit & redemption routes
		if roleCode == "superadmin" || (roleCode == "business_validator" && (menuRoute == "/benefit" || menuRoute == "/company-benefits" || menuRoute == "/redemptions")) {
			c.Next()
			return
		}

		// 2. Get role_id from context
		roleIDVal, exists := c.Get("role_id")
		if !exists || roleIDVal == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: role_id not found in context"))})
			c.Abort()
			return
		}

		var roleID uint
		if f, ok := roleIDVal.(float64); ok {
			roleID = uint(f)
		} else if u, ok := roleIDVal.(uint); ok {
			roleID = u
		} else if u, ok := roleIDVal.(*uint); ok && u != nil {
			roleID = *u
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: invalid role_id format"))})
			c.Abort()
			return
		}

		// 3. Find menu by route
		var menu struct {
			ID uint
		}
		if err := db.Table("administrative.menus").Select("id").Where("route = ? AND is_active = true", menuRoute).First(&menu).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: menu not found or inactive"))})
			c.Abort()
			return
		}

		// 4. Find option by code
		var option struct {
			ID uint
		}
		if err := db.Table("administrative.options").Select("id").Where("code = ?", optionCode).First(&option).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: permission option not found"))})
			c.Abort()
			return
		}

		// 5. Check permission
		var perm struct {
			RoleID uint
		}
		if err := db.Table("administrative.permissions").Select("role_id").
			Where("role_id = ? AND menu_id = ? AND option_id = ?", roleID, menu.ID, option.ID).
			First(&perm).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.TranslateError(c, errors.New("access denied: insufficient permissions for this operation"))})
			c.Abort()
			return
		}

		c.Next()
	}
}
