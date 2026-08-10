package domain

import (
	"time"
)

// --- Estado constants for benefit_redemption ---
const (
	EstadoGenerado = "generado"
	EstadoUsado    = "usado"
	EstadoVencido  = "vencido"
)

// Charset without ambiguous characters (0/O, 1/I/L removed) for human-readable codes.
// This is critical because the employee dictates/writes the code at the business counter.
const CodeCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const CodeLength = 8

// BenefitRedemption represents the administrative.benefit_redemption table.
type BenefitRedemption struct {
	ID               uint                  `gorm:"primaryKey" json:"id"`
	BenefitID        uint                  `gorm:"not null;index:idx_redemption_benefit" json:"benefit_id"`
	UsuarioID        uint                  `gorm:"not null;index:idx_redemption_usuario" json:"usuario_id"`
	CodigoValidacion string                `gorm:"type:varchar(30);not null;uniqueIndex" json:"codigo_validacion"`
	Estado           string                `gorm:"type:varchar(20);not null;default:'generado'" json:"estado"`
	FechaRedencion   time.Time             `gorm:"type:timestamptz;not null;autoCreateTime" json:"fecha_redencion"`
	FechaUso         *time.Time            `gorm:"type:timestamptz" json:"fecha_uso"`
	ValidadoPor      *uint                 `json:"validado_por"`
	Benefit          *BenefitForRedemption `gorm:"foreignKey:BenefitID" json:"benefit,omitempty"`
}

func (BenefitRedemption) TableName() string {
	return "administrative.benefit_redemption"
}

// BenefitForRedemption is a reduced struct for the benefit table,
// used only within this module to avoid coupling with the benefit domain.
type BenefitForRedemption struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"type:varchar(255)" json:"name"`
	Description    string     `gorm:"type:varchar(255)" json:"description"`
	CompanyBenefit *uint      `gorm:"column:company_benefit" json:"company_benefit"`
	Estado         string     `gorm:"column:estado;type:varchar(20)" json:"estado"`
	IsActive       *bool      `gorm:"column:is_active;default:true" json:"is_active"`
	FechaInicio    *time.Time `gorm:"type:date" json:"fecha_inicio"`
	FechaFin       *time.Time `gorm:"type:date" json:"fecha_fin"`
}

func (BenefitForRedemption) TableName() string {
	return "administrative.benefit"
}

// --- Request DTOs ---

type ValidateCodeRequest struct {
	CodigoValidacion string `json:"codigo_validacion" binding:"required"`
}

// --- Response DTOs ---

type RedemptionResponse struct {
	ID               uint       `json:"id"`
	BenefitID        uint       `json:"benefit_id"`
	BenefitName      string     `json:"benefit_name"`
	CodigoValidacion string     `json:"codigo_validacion"`
	Estado           string     `json:"estado"`
	FechaRedencion   time.Time  `json:"fecha_redencion"`
	FechaUso         *time.Time `json:"fecha_uso"`
	ValidadoPor      *uint      `json:"validado_por"`
}

type ValidateCodeResponse struct {
	ID            uint       `json:"id"`
	BenefitName   string     `json:"benefit_name"`
	UsuarioNombre string     `json:"usuario_nombre"`
	Estado        string     `json:"estado"`
	FechaUso      *time.Time `json:"fecha_uso"`
}

// --- Converter functions ---

func ToRedemptionResponse(r *BenefitRedemption) *RedemptionResponse {
	benefitName := ""
	if r.Benefit != nil {
		benefitName = r.Benefit.Name
	}
	return &RedemptionResponse{
		ID:               r.ID,
		BenefitID:        r.BenefitID,
		BenefitName:      benefitName,
		CodigoValidacion: r.CodigoValidacion,
		Estado:           r.Estado,
		FechaRedencion:   r.FechaRedencion,
		FechaUso:         r.FechaUso,
		ValidadoPor:      r.ValidadoPor,
	}
}

func ToRedemptionResponses(redemptions []BenefitRedemption) []*RedemptionResponse {
	responses := make([]*RedemptionResponse, len(redemptions))
	for i, r := range redemptions {
		responses[i] = ToRedemptionResponse(&r)
	}
	return responses
}
