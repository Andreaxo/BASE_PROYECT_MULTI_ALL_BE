package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/benefit_redemption/domain"
	"multicliente-backend/internal/platform/i18n"
)

type RedemptionHandler struct {
	service domain.RedemptionService
}

func NewRedemptionHandler(service domain.RedemptionService) *RedemptionHandler {
	return &RedemptionHandler{service: service}
}

// POST /benefits/:id/redeem — any authenticated user can redeem a benefit
func (h *RedemptionHandler) RedeemBenefit(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid benefit ID")
		return
	}

	redemption, err := h.service.RedeemBenefit(uint(idVal), *userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, redemption)
}

// GET /benefits/mis-redenciones — authenticated user's redemption history
func (h *RedemptionHandler) GetMisRedenciones(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	redemptions, err := h.service.GetMisRedenciones(*userID)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, redemptions)
}

// POST /redemptions/validate — business_validator validates a redemption code
func (h *RedemptionHandler) ValidateCode(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req domain.ValidateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.ValidateCode(req.CodigoValidacion, *userID)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /redemptions — business_validator's company redemption history
func (h *RedemptionHandler) GetRedemptionsByEmpresa(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		i18n.ErrorString(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	redemptions, err := h.service.GetRedemptionsByEmpresa(*userID)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, redemptions)
}

// GET /admin/redemptions — platform admin view of all redemptions
func (h *RedemptionHandler) GetAllRedemptions(c *gin.Context) {
	redemptions, err := h.service.GetAllRedemptions()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, redemptions)
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
