package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/referido/domain"
	"multicliente-backend/internal/platform/i18n"
)

// ReferidoHandler handles HTTP requests for referral operations.
type ReferidoHandler struct {
	service domain.ReferidoService
}

// NewReferidoHandler creates a new ReferidoHandler with the given service.
func NewReferidoHandler(service domain.ReferidoService) *ReferidoHandler {
	return &ReferidoHandler{service: service}
}

// Create handles POST /api/referidos — create a referral invitation.
// The referente is the authenticated user.
func (h *ReferidoHandler) Create(c *gin.Context) {
	var req domain.CreateReferidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	referenteID := getUserIDFromContext(c)
	if referenteID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	response, err := h.service.CreateReferido(&req, *referenteID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetAll handles GET /api/referidos — list all referrals (admin).
func (h *ReferidoHandler) GetAll(c *gin.Context) {
	referidos, err := h.service.GetAllReferidos()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, referidos)
}

// GetByID handles GET /api/referidos/:id
func (h *ReferidoHandler) GetByID(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid referido ID")
		return
	}

	referido, err := h.service.GetReferidoByID(uint(idVal))
	if err != nil {
		i18n.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, referido)
}

// GetMisReferidos handles GET /api/referidos/mis-referidos — list referrals for the authenticated user.
func (h *ReferidoHandler) GetMisReferidos(c *gin.Context) {
	referenteID := getUserIDFromContext(c)
	if referenteID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	referidos, err := h.service.GetReferidosByReferente(*referenteID)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, referidos)
}

// Afiliar handles PUT /api/referidos/:id/afiliar — transition to 'afiliado'.
func (h *ReferidoHandler) Afiliar(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid referido ID")
		return
	}

	response, err := h.service.AfiliarReferido(uint(idVal))
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Recompensa handles PUT /api/referidos/:id/recompensa — grant reward.
func (h *ReferidoHandler) Recompensa(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid referido ID")
		return
	}

	response, err := h.service.OtorgarRecompensa(uint(idVal))
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Delete handles DELETE /api/referidos/:id
func (h *ReferidoHandler) Delete(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid referido ID")
		return
	}

	if err := h.service.DeleteReferido(uint(idVal)); err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "referido eliminado correctamente"})
}

// getUserIDFromContext extracts the authenticated user's uint ID from the Gin context.
func getUserIDFromContext(c *gin.Context) *uint {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	// It could be float64 when decoded from JWT claims map
	if f, ok := userIDVal.(float64); ok {
		u := uint(f)
		return &u
	}

	// Or string
	if s, ok := userIDVal.(string); ok {
		idVal, err := strconv.ParseUint(s, 10, 32)
		if err == nil {
			u := uint(idVal)
			return &u
		}
	}

	return nil
}
