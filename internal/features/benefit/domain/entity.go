package domain

import "time"

type Benefit struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"type:varchar(255);not null" json:"name"`
	Description     string     `gorm:"type:varchar(255)" json:"description"`
	CompanyBenefits *uint      `gorm:"column:company_benefit" json:"company_benefits"`
	Estado          string     `gorm:"type:varchar(20);not null;default:'activo'" json:"estado"`
	IsActive        bool       `gorm:"column:is_active;default:true" json:"is_active"`
	FechaInicio     *time.Time `gorm:"type:date" json:"fecha_inicio"`
	FechaFin        *time.Time `gorm:"type:date" json:"fecha_fin"`
}

func (Benefit) TableName() string {
	return "administrative.benefit"
}

// DTOs
type CreateBenefitRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	CompanyBenefits *uint  `json:"company_benefits"`
}

type UpdateBenefitRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	CompanyBenefits *uint   `json:"company_benefits"`
}
