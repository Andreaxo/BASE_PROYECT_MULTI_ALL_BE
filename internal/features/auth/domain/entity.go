package domain

// LoginRequest is the payload for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token                  string   `json:"token"`
	User                   UserInfo `json:"user"`
	SessionDurationSeconds int      `json:"session_duration_seconds"`
}

// UserInfo contains the basic user data included in the login response.
type UserInfo struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    *uint  `json:"role_id"`
	RoleCode  string `json:"role_code"`
}

// RegisterRequest is the payload for the public registration endpoint.
type RegisterRequest struct {
	Email          string  `json:"email" binding:"required,email"`
	Password       string  `json:"password" binding:"required,min=10"`
	FirstName      string  `json:"first_name" binding:"required"`
	LastName       string  `json:"last_name" binding:"required"`
	RefCode        *string `json:"ref_code"`
	CodigoEmpresa  *string `json:"codigo_empresa"`
}

// RegisterResponse is returned after a successful registration.
type RegisterResponse struct {
	Token                  string   `json:"token"`
	User                   UserInfo `json:"user"`
	SessionDurationSeconds int      `json:"session_duration_seconds"`
	ReferidoCreated        bool     `json:"referido_created"`
}

// ChangePasswordRequest represents the payload to change own password in profile
type ChangePasswordRequest struct {
	CurrentPassword        string `json:"current_password" binding:"required"`
	ConfirmCurrentPassword string `json:"confirm_current_password" binding:"required"`
	NewPassword            string `json:"new_password" binding:"required,min=10"`
}

// UpdateProfileRequest represents the payload to update own profile details
type UpdateProfileRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// OlvidePasswordRequest is the payload for requesting a password reset email.
type OlvidePasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest is the payload for setting a new password with a token.
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=10"`
}

// RegistroAsistidoRequest is the payload for operator-assisted user registration.
type RegistroAsistidoRequest struct {
	Email         string  `json:"email" binding:"required,email"`
	FirstName     string  `json:"first_name" binding:"required"`
	LastName      string  `json:"last_name" binding:"required"`
	RefCode       *string `json:"ref_code"`
	CodigoEmpresa *string `json:"codigo_empresa"`
	EmpresaID     *uint   `json:"empresa_id"`
}

// RegistroAsistidoResponse is returned after assisted registration.
type RegistroAsistidoResponse struct {
	User            UserInfo `json:"user"`
	ReferidoCreated bool     `json:"referido_created"`
	EmailSent       bool     `json:"email_sent"`
	Message         string   `json:"message"`
}

