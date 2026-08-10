package domain

// UserRepository defines the secondary port for user data persistence.
type UserRepository interface {
	Create(user *User) error
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByCodeRefer(code string) (*User, error)
	FindAll() ([]User, error)
	Update(user *User) error
	Delete(id uint) error
}

// UserService defines the primary port for user business operations.
type UserService interface {
	CreateUser(req *CreateUserRequest, createdBy *uint) (*UserResponse, error)
	GetUserByID(id uint) (*UserResponse, error)
	GetAllUsers() ([]*UserResponse, error)
	UpdateUser(id uint, req *UpdateUserRequest, updatedBy *uint) (*UserResponse, error)
	DeleteUser(id uint) error
}
