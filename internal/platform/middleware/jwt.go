package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"multicliente-backend/internal/platform/i18n"
)

// JWTAuth returns a Gin middleware that validates Bearer JWT tokens.
// It extracts user_id and email from claims and sets them in the context.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				tokenString = parts[1]
			}
		}

		// Try loading from cookie if header is empty or has invalid format
		if tokenString == "" {
			if cookieToken, err := c.Cookie("token"); err == nil {
				tokenString = cookieToken
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("authorization token is required"))})
			c.Abort()
			return
		}

		if GlobalTokenBlacklist.IsRevoked(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "La sesión ha sido cerrada o el token fue revocado."})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("invalid or expired token"))})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, errors.New("invalid token claims"))})
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("role", claims["role"])
		c.Set("role_id", claims["role_id"])

		c.Next()
	}
}
