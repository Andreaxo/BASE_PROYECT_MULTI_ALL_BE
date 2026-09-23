package domain

import "time"

// --- Constants ---

const (
	MembresiaInactiva  = "inactiva"
	MembresiaActiva    = "activa"
	MembresiaVencida   = "vencida"
	MembresiaCancelada = "cancelada"

	PagoTipoInicial    = "inicial"
	PagoTipoRenovacion = "renovacion"

	PagoEstadoPendiente  = "pendiente"
	PagoEstadoAprobado   = "aprobado"
	PagoEstadoRechazado  = "rechazado"
)

// --- Entities ---

// Membresia represents the administrative.membresia table.
type Membresia struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	UsuarioID           uint       `gorm:"uniqueIndex;not null" json:"usuario_id"`
	Estado              string     `gorm:"type:varchar(20);not null;default:'inactiva'" json:"estado"`
	FechaInicio         *time.Time `gorm:"type:timestamptz" json:"fecha_inicio"`
	FechaFin            *time.Time `gorm:"type:timestamptz" json:"fecha_fin"`
	RenovacionAutomatica bool      `gorm:"not null;default:true" json:"renovacion_automatica"`
	MetodoPagoToken     *string    `gorm:"type:varchar(150)" json:"metodo_pago_token,omitempty"`
	CreateAt            time.Time  `gorm:"type:timestamptz;not null;autoCreateTime" json:"create_at"`
}

func (Membresia) TableName() string {
	return "administrative.membresia"
}

// PagoMembresia represents the administrative.pago_membresia table.
type PagoMembresia struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	MembresiaID       uint      `gorm:"not null;index:idx_pago_membresia" json:"membresia_id"`
	Monto             float64   `gorm:"type:decimal(10,2);not null" json:"monto"`
	Moneda            string    `gorm:"type:varchar(10);not null;default:'COP'" json:"moneda"`
	Tipo              string    `gorm:"type:varchar(20);not null;default:'inicial'" json:"tipo"`
	ReferenciaPasarela *string  `gorm:"type:varchar(100)" json:"referencia_pasarela"`
	Estado            string    `gorm:"type:varchar(20);not null;default:'pendiente'" json:"estado"`
	FechaPago         time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"fecha_pago"`
}

func (PagoMembresia) TableName() string {
	return "administrative.pago_membresia"
}

// --- DTOs ---

// IniciarPagoRequest is the payload to start a membership payment flow.
type IniciarPagoRequest struct {
	RedirectURL string `json:"redirect_url"`
}

// IniciarPagoResponse is returned with the Wompi checkout URL.
type IniciarPagoResponse struct {
	TransactionID string `json:"transaction_id"`
	CheckoutURL   string `json:"checkout_url"`
	Referencia    string `json:"referencia"`
}

// MembresiaResponse is the public representation of a membership.
type MembresiaResponse struct {
	ID                   uint       `json:"id"`
	UsuarioID            uint       `json:"usuario_id"`
	Estado               string     `json:"estado"`
	FechaInicio          *time.Time `json:"fecha_inicio"`
	FechaFin             *time.Time `json:"fecha_fin"`
	RenovacionAutomatica bool       `json:"renovacion_automatica"`
	TieneMetodoPago      bool       `json:"tiene_metodo_pago"`
	CreateAt             time.Time  `json:"create_at"`
	UltimoPagoEstado     *string    `json:"ultimo_pago_estado,omitempty"`
}

// ToMembresiaResponse converts a Membresia entity to a MembresiaResponse DTO.
func ToMembresiaResponse(m *Membresia) *MembresiaResponse {
	return &MembresiaResponse{
		ID:                   m.ID,
		UsuarioID:            m.UsuarioID,
		Estado:               m.Estado,
		FechaInicio:          m.FechaInicio,
		FechaFin:             m.FechaFin,
		RenovacionAutomatica: m.RenovacionAutomatica,
		TieneMetodoPago:      m.MetodoPagoToken != nil && *m.MetodoPagoToken != "",
		CreateAt:             m.CreateAt,
	}
}

// PagoMembresiaResponse is the public representation of a payment.
type PagoMembresiaResponse struct {
	ID                 uint      `json:"id"`
	MembresiaID        uint      `json:"membresia_id"`
	Monto              float64   `json:"monto"`
	Moneda             string    `json:"moneda"`
	Tipo               string    `json:"tipo"`
	ReferenciaPasarela *string   `json:"referencia_pasarela"`
	Estado             string    `json:"estado"`
	FechaPago          time.Time `json:"fecha_pago"`
}
