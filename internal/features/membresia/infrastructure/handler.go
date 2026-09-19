package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/membresia/domain"
	"multicliente-backend/internal/platform/i18n"
)

// MembresiaHandler handles HTTP requests for membership operations.
type MembresiaHandler struct {
	service domain.MembresiaService
}

// NewMembresiaHandler creates a new MembresiaHandler.
func NewMembresiaHandler(service domain.MembresiaService) *MembresiaHandler {
	return &MembresiaHandler{service: service}
}

// IniciarPago handles POST /api/membresia/iniciar-pago
func (h *MembresiaHandler) IniciarPago(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req domain.IniciarPagoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	response, err := h.service.IniciarPago(*userID, &req)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMiMembresia handles GET /api/membresia/mi-membresia
func (h *MembresiaHandler) GetMiMembresia(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	response, err := h.service.GetMiMembresia(*userID)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelarRenovacion handles PUT /api/membresia/cancelar-renovacion
func (h *MembresiaHandler) CancelarRenovacion(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	response, err := h.service.CancelarRenovacion(*userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAll handles GET /api/membresia/admin/todas (admin endpoint)
func (h *MembresiaHandler) GetAll(c *gin.Context) {
	membresias, err := h.service.GetAllMembresias()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, membresias)
}

// SimularPago handles POST /api/membresia/simular-pago
func (h *MembresiaHandler) SimularPago(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		req.Status = "APPROVED"
	}

	res, err := h.service.SimularPago(*userID, req.Status)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// getUserIDFromContext extracts the authenticated user's uint ID from the Gin context.
func getUserIDFromContext(c *gin.Context) *uint {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	if f, ok := userIDVal.(float64); ok {
		u := uint(f)
		return &u
	}

	if s, ok := userIDVal.(string); ok {
		idVal, err := strconv.ParseUint(s, 10, 32)
		if err == nil {
			u := uint(idVal)
			return &u
		}
	}

	return nil
}
