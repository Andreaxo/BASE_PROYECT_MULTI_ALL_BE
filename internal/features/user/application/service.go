package application

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	companyDomain "multicliente-backend/internal/features/company/domain"
	"multicliente-backend/internal/features/user/domain"
)

type userService struct {
	repo domain.UserRepository
}

// NewUserService creates a new UserService with the given repository.
func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(req *domain.CreateUserRequest, createdBy *uint) (*domain.UserResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Map companies
	companies := make([]companyDomain.Company, len(req.CompanyIDs))
	for i, cid := range req.CompanyIDs {
		companies[i] = companyDomain.Company{ID: cid}
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		CreateBy:  createdBy,
		RoleID:    req.RoleID,
		Companies: companies,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// Fetch fully loaded user relations for the response
	fullUser, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.ToUserResponse(fullUser), nil
}

func (s *userService) GetUserByID(id uint) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return domain.ToUserResponse(user), nil
}

func (s *userService) GetAllUsers() ([]*domain.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return domain.ToUserResponses(users), nil
}

func (s *userService) UpdateUser(id uint, req *domain.UpdateUserRequest, updatedBy *uint) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Email != nil {
		existing, _ := s.repo.FindByEmail(*req.Email)
		if existing != nil && existing.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	if req.CompanyIDs != nil {
		companies := make([]companyDomain.Company, len(req.CompanyIDs))
		for i, cid := range req.CompanyIDs {
			companies[i] = companyDomain.Company{ID: cid}
		}
		user.Companies = companies
	}

	user.UpdateBy = updatedBy

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	// Fetch fully loaded updated user relations
	fullUser, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.ToUserResponse(fullUser), nil
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(id)
}
