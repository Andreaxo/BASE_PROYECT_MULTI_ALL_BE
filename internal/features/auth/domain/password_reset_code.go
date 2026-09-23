package domain

import "time"

// PasswordResetCode representa un código numérico temporal de 6 dígitos para recuperación de contraseña.
type PasswordResetCode struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UsuarioID        uint      `gorm:"not null;index:idx_password_reset_code_usuario" json:"usuario_id"`
	CodigoHash       string    `gorm:"type:varchar(255);not null" json:"codigo_hash"`
	IntentosFallidos int       `gorm:"not null;default:0" json:"intentos_fallidos"`
	Usado            bool      `gorm:"not null;default:false" json:"usado"`
	ExpiraAt         time.Time `gorm:"not null;type:timestamptz" json:"expira_at"`
	CreateAt         time.Time `gorm:"not null;type:timestamptz;default:now()" json:"create_at"`
}

// TableName especifica el nombre de la tabla en el esquema administrative.
func (PasswordResetCode) TableName() string {
	return "administrative.password_reset_code"
}

// VerificarCodigoRequest es la carga útil para validar el código de 6 dígitos recibido por correo.
type VerificarCodigoRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Codigo string `json:"codigo" binding:"required,len=6"`
}

// VerificarCodigoResponse es la respuesta cuando el código de verificación es correcto.
type VerificarCodigoResponse struct {
	ResetToken string `json:"reset_token"`
	Message    string `json:"message"`
}
