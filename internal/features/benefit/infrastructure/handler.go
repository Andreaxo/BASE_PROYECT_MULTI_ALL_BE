package infrastructure

import (
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"multicliente-backend/internal/features/benefit/domain"
	"multicliente-backend/internal/platform/i18n"

	"github.com/gin-gonic/gin"
)

type BenefitHandler struct {
	service domain.BenefitService
	db      *gorm.DB
}

func NewBenefitHandler(service domain.BenefitService, db *gorm.DB) *BenefitHandler {
	return &BenefitHandler{service: service, db: db}
}

func (h *BenefitHandler) getUserEmpresaAndRole(c *gin.Context) (*uint, string) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil, ""
	}
	var userID uint
	if f, ok := userIDVal.(float64); ok {
		userID = uint(f)
	} else if s, ok := userIDVal.(string); ok {
		if idVal, err := strconv.ParseUint(s, 10, 32); err == nil {
			userID = uint(idVal)
		}
	}
	if userID == 0 {
		return nil, ""
	}

	var user struct {
		RoleCode  string `gorm:"column:role_code"`
		EmpresaID *uint  `gorm:"column:empresa_id"`
	}
	h.db.Table("administrative.users u").
		Select("r.code as role_code, u.empresa_id").
		Joins("JOIN administrative.roles r ON r.id = u.role_id").
		Where("u.id = ?", userID).
		Scan(&user)

	if user.EmpresaID == nil || *user.EmpresaID == 0 {
		var userComp struct {
			CompanyID uint `gorm:"column:company_id"`
		}
		if err := h.db.Table("administrative.user_companies").
			Select("company_id").
			Where("user_id = ?", userID).
			Order("company_id ASC").
			Limit(1).
			Scan(&userComp).Error; err == nil && userComp.CompanyID > 0 {
			user.EmpresaID = &userComp.CompanyID
		}
	}

	return user.EmpresaID, user.RoleCode
}

func (h *BenefitHandler) Create(c *gin.Context) {
	var req domain.CreateBenefitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	empresaID, roleCode := h.getUserEmpresaAndRole(c)
	if roleCode != "superadmin" && roleCode != "admin" {
		if empresaID == nil || *empresaID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tu usuario no tiene una empresa vinculada para crear beneficios"})
			return
		}
		// Una empresa solo puede crear beneficios para su propia empresa
		req.CompanyBenefits = empresaID
	}

	benefit, err := h.service.CreateBenefit(&req)
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, benefit)
}

func (h *BenefitHandler) GetAll(c *gin.Context) {
	benefit, err := h.service.GetAllBenefits()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, benefit)
}

func (h *BenefitHandler) GetMyCompanyBenefits(c *gin.Context) {
	empresaID, roleCode := h.getUserEmpresaAndRole(c)
	if roleCode != "superadmin" && roleCode != "admin" {
		if empresaID != nil {
			benefits, err := h.service.GetBenefitsByCompany(*empresaID)
			if err != nil {
				i18n.Error(c, http.StatusInternalServerError, err)
				return
			}
			c.JSON(http.StatusOK, benefits)
			return
		}
		c.JSON(http.StatusOK, []domain.Benefit{})
		return
	}

	// Fallback to all benefits if admin
	benefits, err := h.service.GetAllBenefits()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, benefits)
}

func (h *BenefitHandler) GetBenefitByID(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid benefit ID")
		return
	}
	id := uint(idVal)

	benefit, err := h.service.GetBenefitByID(id)
	if err != nil {
		i18n.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, benefit)
}

func (h *BenefitHandler) Update(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid benefit ID")
		return
	}
	id := uint(idVal)

	empresaID, roleCode := h.getUserEmpresaAndRole(c)
	if roleCode != "superadmin" && roleCode != "admin" {
		existing, err := h.service.GetBenefitByID(id)
		if err != nil {
			i18n.Error(c, http.StatusNotFound, err)
			return
		}
		if empresaID == nil || existing.CompanyBenefits == nil || *existing.CompanyBenefits != *empresaID {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para modificar beneficios de otra empresa"})
			return
		}
	}

	var req domain.UpdateBenefitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	if roleCode != "superadmin" && roleCode != "admin" {
		req.CompanyBenefits = empresaID
	}

	benefit, err := h.service.UpdateBenefit(id, &req)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, benefit)
}

func (h *BenefitHandler) Delete(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid benefit ID")
		return
	}
	id := uint(idVal)

	empresaID, roleCode := h.getUserEmpresaAndRole(c)
	if roleCode != "superadmin" && roleCode != "admin" {
		existing, err := h.service.GetBenefitByID(id)
		if err != nil {
			i18n.Error(c, http.StatusNotFound, err)
			return
		}
		if empresaID == nil || existing.CompanyBenefits == nil || *existing.CompanyBenefits != *empresaID {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para eliminar beneficios de otra empresa"})
			return
		}
	}

	if err := h.service.DeleteBenefit(id); err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "benefit deleted successfully"})
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
