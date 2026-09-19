package domain

import (
	companyDomain "multicliente-backend/internal/features/company/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
)

// AuthService defines the primary port for authentication operations.
type AuthService interface {
	Login(req *LoginRequest) (*LoginResponse, error)
	Register(req *RegisterRequest) (*RegisterResponse, error)
	GetProfile(userID uint) (*userDomain.User, error)
	GetMyCodeRefer(userID uint) (string, error)
	ChangePassword(userID uint, req *ChangePasswordRequest) error
	UpdateProfile(userID uint, req *UpdateProfileRequest) (*userDomain.User, error)
	ValidarCodigoEmpresa(codigo string) (*companyDomain.Company, error)
}

