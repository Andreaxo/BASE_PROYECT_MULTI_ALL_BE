package infrastructure

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/notificacion/domain"
	"multicliente-backend/internal/platform/i18n"
)

// NotificacionHandler maneja las peticiones HTTP para notificaciones.
type NotificacionHandler struct {
	service domain.NotificacionService
}

// NewNotificacionHandler crea un nuevo handler de notificaciones.
func NewNotificacionHandler(service domain.NotificacionService) *NotificacionHandler {
	return &NotificacionHandler{service: service}
}

func getUserIDFromContext(c *gin.Context) (uint, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("usuario no autenticado")
	}
	if f, ok := val.(float64); ok {
		return uint(f), nil
	}
	if u, ok := val.(uint); ok {
		return u, nil
	}
	return 0, errors.New("formato de id de usuario inválido")
}

// GetNotificaciones handles GET /api/notificaciones
func (h *NotificacionHandler) GetNotificaciones(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	notificaciones, err := h.service.GetNotificaciones(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notificaciones})
}

// GetCountNoLeidas handles GET /api/notificaciones/no-leidas/count
func (h *NotificacionHandler) GetCountNoLeidas(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	count, err := h.service.GetCountNoLeidas(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, domain.NotificacionCountResponse{Count: count})
}

// MarcarLeida handles PUT /api/notificaciones/:id/leido
func (h *NotificacionHandler) MarcarLeida(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	idParam := c.Param("id")
	notifID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de notificación inválido"})
		return
	}

	if err := h.service.MarcarLeida(uint(notifID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notificación marcada como leída"})
}

// MarcarTodasLeidas handles PUT /api/notificaciones/leer-todas
func (h *NotificacionHandler) MarcarTodasLeidas(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	if err := h.service.MarcarTodasLeidas(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.TranslateError(c, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todas las notificaciones marcadas como leídas"})
}
