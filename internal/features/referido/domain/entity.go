package domain

import (
	"time"

	userDomain "multicliente-backend/internal/features/user/domain"
)

// EstadoReferido defines the allowed states for a referral lifecycle.
const (
	EstadoPendiente  = "pendiente"
	EstadoRegistrado = "registrado"
	EstadoAfiliado   = "afiliado"
)

// Referido represents the administrative.referido table.
// It models the referral EVENT (not just the referred person).
// - UsuarioReferenteID: who shared their code and referred.
// - UsuarioReferidoID: who was referred/invited.
// Personal data for both actors is obtained via JOIN to users (never duplicated here).
type Referido struct {
	ID                 uint             `gorm:"primaryKey" json:"id"`
	UsuarioReferenteID uint             `gorm:"not null;index:idx_referido_referente" json:"usuario_referente_id"`
	UsuarioReferidoID  *uint            `gorm:"index:idx_referido_referido" json:"usuario_referido_id"`
	EmailReferido      string           `gorm:"type:varchar(150);not null;index:idx_referido_email" json:"email_referido"`
	NombreReferido     *string          `gorm:"type:varchar(150)" json:"nombre_referido"`
	Estado             string           `gorm:"type:varchar(20);not null;default:'pendiente'" json:"estado"`
	FechaReferido      time.Time        `gorm:"type:timestamptz;not null;autoCreateTime" json:"fecha_referido"`
	FechaConversion    *time.Time       `gorm:"type:timestamptz" json:"fecha_conversion"`
	RecompensaOtorgada bool             `gorm:"not null;default:false" json:"recompensa_otorgada"`
	Referente          *userDomain.User `gorm:"foreignKey:UsuarioReferenteID" json:"referente,omitempty"`
	Referido_          *userDomain.User `gorm:"foreignKey:UsuarioReferidoID" json:"referido_usuario,omitempty"`
}

// TableName places the table inside the administrative schema.
func (Referido) TableName() string {
	return "administrative.referido"
}

// --- DTOs ---

// CreateReferidoRequest is the payload to create a new referral invitation.
type CreateReferidoRequest struct {
	EmailReferido  string  `json:"email_referido" binding:"required,email"`
	NombreReferido *string `json:"nombre_referido"`
}

// ReferidoResponse is the public representation of a referral record.
type ReferidoResponse struct {
	ID                 uint       `json:"id"`
	UsuarioReferenteID uint       `json:"usuario_referente_id"`
	NombreReferente    string     `json:"nombre_referente"`
	UsuarioReferidoID  *uint      `json:"usuario_referido_id"`
	NombreReferidoUser string     `json:"nombre_referido_user"`
	EmailReferido      string     `json:"email_referido"`
	NombreReferido     *string    `json:"nombre_referido"`
	Estado             string     `json:"estado"`
	FechaReferido      time.Time  `json:"fecha_referido"`
	FechaConversion    *time.Time `json:"fecha_conversion"`
	RecompensaOtorgada bool       `json:"recompensa_otorgada"`
}

// ToReferidoResponse converts a Referido entity to a ReferidoResponse DTO.
func ToReferidoResponse(r *Referido) *ReferidoResponse {
	nombreReferente := ""
	if r.Referente != nil {
		nombreReferente = r.Referente.FirstName + " " + r.Referente.LastName
	}
	nombreReferidoUser := ""
	if r.Referido_ != nil {
		nombreReferidoUser = r.Referido_.FirstName + " " + r.Referido_.LastName
	}
	return &ReferidoResponse{
		ID:                 r.ID,
		UsuarioReferenteID: r.UsuarioReferenteID,
		NombreReferente:    nombreReferente,
		UsuarioReferidoID:  r.UsuarioReferidoID,
		NombreReferidoUser: nombreReferidoUser,
		EmailReferido:      r.EmailReferido,
		NombreReferido:     r.NombreReferido,
		Estado:             r.Estado,
		FechaReferido:      r.FechaReferido,
		FechaConversion:    r.FechaConversion,
		RecompensaOtorgada: r.RecompensaOtorgada,
	}
}

// ToReferidoResponses converts a slice of Referido entities to ReferidoResponse DTOs.
func ToReferidoResponses(referidos []Referido) []*ReferidoResponse {
	responses := make([]*ReferidoResponse, len(referidos))
	for i, r := range referidos {
		responses[i] = ToReferidoResponse(&r)
	}
	return responses
}
