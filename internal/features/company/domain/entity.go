package domain

import (
	"time"
)

type Company struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Name              string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	IsActive          bool       `gorm:"default:true" json:"is_active"`
	PhotoURL          string     `gorm:"type:varchar(255);default:''" json:"photo_url"`
	CreateBy          *uint      `json:"create_by"`
	CreateAt          time.Time  `gorm:"autoCreateTime" json:"create_at"`
	UpdateBy          *uint      `json:"update_by"`
	UpdateAt          time.Time  `gorm:"autoUpdateTime" json:"update_at"`
	CreateByName      string     `gorm:"-" json:"create_by_name"`
	UpdateByName      string     `gorm:"-" json:"update_by_name"`
	NIT               *uint      `gorm:"type:bigint" json:"nit"`
	RazonSocial       string     `gorm:"type:varchar(255);default:''" json:"razon_social"`
	SuscripcionEstado string     `gorm:"type:varchar(20);not null;default:'prueba'" json:"suscripcion_estado"`
	FechaFinPrueba    *time.Time `gorm:"type:timestamptz" json:"fecha_fin_prueba"`
	CodigoEmpresa     *string    `gorm:"type:varchar(10);uniqueIndex" json:"codigo_empresa"`
}

func (Company) TableName() string {
	return "administrative.companies"
}

// DTOs
type CreateCompanyRequest struct {
	Name               string `json:"name" binding:"required"`
	PhotoURL           string `json:"photo_url"`
	NIT                *uint  `json:"nit"`
	RazonSocial        string `json:"razon_social"`
	ValidatorEmail     string `json:"validator_email"`
	ValidatorPassword  string `json:"validator_password"`
	ValidatorFirstName string `json:"validator_first_name"`
	ValidatorLastName  string `json:"validator_last_name"`
}

type UpdateCompanyRequest struct {
	Name        *string `json:"name"`
	IsActive    *bool   `json:"is_active"`
	PhotoURL    *string `json:"photo_url"`
	NIT         *uint   `json:"nit"`
	RazonSocial *string `json:"razon_social"`
}
