package domain

import (
	"time"

	companyDomain "multicliente-backend/internal/features/company/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
)

// User represents the users table with audit fields.
type User struct {
	ID           uint                    `gorm:"primaryKey" json:"id"`
	Email        string                  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password     string                  `gorm:"type:varchar(255);not null" json:"-"`
	FirstName    string                  `gorm:"type:varchar(100)" json:"first_name"`
	LastName     string                  `gorm:"type:varchar(100)" json:"last_name"`
	IsActive     bool                    `gorm:"default:true" json:"is_active"`
	CreateBy     *uint                   `json:"create_by"`
	CreateAt     time.Time               `gorm:"autoCreateTime" json:"create_at"`
	UpdateBy     *uint                   `json:"update_by"`
	UpdateAt     time.Time               `gorm:"autoUpdateTime" json:"update_at"`
	CreateByName string                  `gorm:"-" json:"create_by_name"`
	UpdateByName string                  `gorm:"-" json:"update_by_name"`
	RoleID       *uint                   `json:"role_id"`
	Role         *roleDomain.Role        `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Companies    []companyDomain.Company `gorm:"many2many:administrative.user_companies;" json:"companies,omitempty"`
	EmpresaID    uint                    `gorm:"-" json:"empresa_id"`
	CodeRefer    string                  `gorm:"type:varchar(10);uniqueIndex" json:"code_refer"`
}

// TableName overrides the default GORM table name mapping to place the table inside the administrative schema.
func (User) TableName() string {
	return "administrative.users"
}

// --- DTOs ---

// CreateUserRequest is the payload for creating a new user.
type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	RoleID     *uint  `json:"role_id"`
	CompanyIDs []uint `json:"company_ids"`
}

// UpdateUserRequest is the payload for updating an existing user.
// All fields are optional (pointer types).
type UpdateUserRequest struct {
	Email      *string `json:"email" binding:"omitempty,email"`
	Password   *string `json:"password" binding:"omitempty,min=6"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	IsActive   *bool   `json:"is_active"`
	RoleID     *uint   `json:"role_id"`
	CompanyIDs []uint  `json:"company_ids"`
}

// UserResponse is the public representation of a user (no password).
type UserResponse struct {
	ID           uint                    `json:"id"`
	Email        string                  `json:"email"`
	FirstName    string                  `json:"first_name"`
	LastName     string                  `json:"last_name"`
	IsActive     bool                    `json:"is_active"`
	CreateBy     *uint                   `json:"create_by"`
	CreateAt     time.Time               `json:"create_at"`
	UpdateBy     *uint                   `json:"update_by"`
	UpdateAt     time.Time               `json:"update_at"`
	CreateByName string                  `json:"create_by_name"`
	UpdateByName string                  `json:"update_by_name"`
	RoleID       *uint                   `json:"role_id"`
	RoleCode     string                  `json:"role_code"`
	Companies    []companyDomain.Company `json:"companies,omitempty"`
	CodeRefer    string                  `json:"code_refer"`
}

// ToUserResponse converts a User entity to a UserResponse DTO.
func ToUserResponse(u *User) *UserResponse {
	roleCode := ""
	if u.Role != nil {
		roleCode = u.Role.Code
	}
	return &UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		IsActive:     u.IsActive,
		CreateBy:     u.CreateBy,
		CreateAt:     u.CreateAt,
		UpdateBy:     u.UpdateBy,
		UpdateAt:     u.UpdateAt,
		CreateByName: u.CreateByName,
		UpdateByName: u.UpdateByName,
		RoleID:       u.RoleID,
		RoleCode:     roleCode,
		Companies:    u.Companies,
		CodeRefer:    u.CodeRefer,
	}
}

// ToUserResponses converts a slice of User entities to UserResponse DTOs.
func ToUserResponses(users []User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, u := range users {
		responses[i] = ToUserResponse(&u)
	}
	return responses
}
