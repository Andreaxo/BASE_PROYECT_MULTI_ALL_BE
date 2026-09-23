package domain

import "time"

const (
	TipoTokenRecuperacion = "recuperacion"
	TipoTokenActivacion   = "activacion"
)

// PasswordResetToken representa un token temporal de un solo uso para recuperación de contraseña o activación de cuenta.
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UsuarioID uint      `gorm:"not null;index" json:"usuario_id"`
	Token     string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"token"`
	ExpiraAt  time.Time `gorm:"not null;index" json:"expira_at"`
	Usado     bool      `gorm:"default:false;index" json:"usado"`
	Tipo      string    `gorm:"type:varchar(30);not null;default:'recuperacion'" json:"tipo"`
	CreateAt  time.Time `gorm:"autoCreateTime" json:"create_at"`
}

// TableName especifica el esquema administrative para la tabla en PostgreSQL.
func (PasswordResetToken) TableName() string {
	return "administrative.password_reset_tokens"
}
