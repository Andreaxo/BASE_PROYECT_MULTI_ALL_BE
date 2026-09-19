package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"multicliente-backend/internal/features/company/domain"
	"multicliente-backend/internal/platform/i18n"
)

type CompanyHandler struct {
	service domain.CompanyService
}

func NewCompanyHandler(service domain.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var req domain.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	createdBy := getUserIDFromContext(c)

	company, err := h.service.CreateCompany(&req, createdBy)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, company)
}

func (h *CompanyHandler) GetAll(c *gin.Context) {
	companies, err := h.service.GetAllCompanies()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, companies)
}

func (h *CompanyHandler) GetByID(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid company ID")
		return
	}
	id := uint(idVal)

	company, err := h.service.GetCompanyByID(id)
	if err != nil {
		i18n.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, company)
}

func (h *CompanyHandler) Update(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid company ID")
		return
	}
	id := uint(idVal)

	var req domain.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	updatedBy := getUserIDFromContext(c)

	company, err := h.service.UpdateCompany(id, &req, updatedBy)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, company)
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid company ID")
		return
	}
	id := uint(idVal)

	if err := h.service.DeleteCompany(id); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company deleted successfully"})
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
