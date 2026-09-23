package domain

import "time"

const (
	TipoPagoAprobado  = "pago_membresia_aprobado"
	TipoPagoRechazado = "pago_membresia_rechazado"
)

// Notificacion representa una notificación interna en el sistema.
type Notificacion struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UsuarioID      uint      `gorm:"not null;index" json:"usuario_id"`
	Tipo           string    `gorm:"type:varchar(50);not null" json:"tipo"`
	Titulo         string    `gorm:"type:varchar(200);not null" json:"titulo"`
	Mensaje        string    `gorm:"type:text;not null" json:"mensaje"`
	Leido          bool      `gorm:"default:false;index" json:"leido"`
	ReferenciaID   *uint     `json:"referencia_id,omitempty"`
	ReferenciaTipo *string   `gorm:"type:varchar(50)" json:"referencia_tipo,omitempty"`
	CreateAt       time.Time `gorm:"autoCreateTime" json:"create_at"`
}

// TableName define el nombre de tabla con esquema administrative.
func (Notificacion) TableName() string {
	return "administrative.notificaciones"
}

// NotificacionCountResponse representa la respuesta para el conteo de notificaciones no leídas.
type NotificacionCountResponse struct {
	Count int64 `json:"count"`
}
