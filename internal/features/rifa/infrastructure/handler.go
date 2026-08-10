package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/rifa/domain"
	"multicliente-backend/internal/platform/i18n"
)

type RifaHandler struct {
	service domain.RifaService
}

func NewRifaHandler(service domain.RifaService) *RifaHandler {
	return &RifaHandler{service: service}
}

// GET /api/rifas/activa
func (h *RifaHandler) GetActiva(c *gin.Context) {
	rifa, err := h.service.GetRifaActiva()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rifa)
}

// GET /api/rifas/mis-participaciones
func (h *RifaHandler) GetMisParticipaciones(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	participaciones, err := h.service.GetMisParticipaciones(*userID)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, participaciones)
}

// GET /api/rifas/historial
func (h *RifaHandler) GetHistorial(c *gin.Context) {
	historial, err := h.service.GetHistorial()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, historial)
}

// GET /api/rifas
func (h *RifaHandler) GetAll(c *gin.Context) {
	rifas, err := h.service.GetAllRifas()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rifas)
}

// GET /api/rifas/:id
func (h *RifaHandler) GetByID(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	rifa, err := h.service.GetRifaByID(uint(idVal))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rifa)
}

// POST /api/rifas
func (h *RifaHandler) Create(c *gin.Context) {
	var req domain.CreateRifaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.CreateRifa(&req, userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, rifa)
}

// PUT /api/rifas/:id
func (h *RifaHandler) Update(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	var req domain.UpdateRifaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.UpdateRifa(uint(idVal), &req, userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rifa)
}

// PUT /api/rifas/:id/activar
func (h *RifaHandler) Activar(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.ActivarRifa(uint(idVal), userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rifa)
}

// PUT /api/rifas/:id/cerrar
func (h *RifaHandler) Cerrar(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.CerrarRifa(uint(idVal), userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rifa)
}

// POST /api/rifas/:id/sortear
func (h *RifaHandler) Sortear(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.SortearRifa(uint(idVal), userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rifa)
}

// PUT /api/rifas/:id/cancelar
func (h *RifaHandler) Cancelar(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	userID := getUserIDFromContext(c)
	rifa, err := h.service.CancelarRifa(uint(idVal), userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, rifa)
}

// GET /api/rifas/:id/participaciones
func (h *RifaHandler) GetParticipaciones(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	participaciones, err := h.service.GetParticipacionesByRifa(uint(idVal))
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, participaciones)
}

// POST /api/rifas/:id/participaciones
func (h *RifaHandler) AgregarParticipacion(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	var req domain.AgregarParticipacionManualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	participacion, err := h.service.AgregarParticipacion(uint(idVal), req.UsuarioID, domain.OrigenManual)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, participacion)
}

// POST /api/rifas/:id/participar — authenticated user joins an active raffle
func (h *RifaHandler) Participar(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	participacion, err := h.service.AgregarParticipacion(uint(idVal), *userID, domain.OrigenManual)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, participacion)
}

// GET /api/rifas/:id/ganadores
func (h *RifaHandler) GetGanadores(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	ganadores, err := h.service.GetGanadoresByRifa(uint(idVal))
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ganadores)
}

// POST /api/rifas/:id/ganadores
func (h *RifaHandler) RegistrarGanador(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid rifa ID")
		return
	}

	var req domain.RegistrarGanadorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	ganador, err := h.service.RegistrarGanador(uint(idVal), req.ParticipacionID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, ganador)
}

// PUT /api/rifas/ganadores/:id/entrega
func (h *RifaHandler) ActualizarEntrega(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid ganador ID")
		return
	}

	var req domain.ActualizarEntregaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	ganador, err := h.service.ActualizarEstadoEntrega(uint(idVal), req.EstadoEntrega)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, ganador)
}

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
