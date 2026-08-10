package infrastructure

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/company/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/database"
)

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) domain.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) populateAudits(companies []*domain.Company) {
	var userIDs []uint
	for _, c := range companies {
		if c.CreateBy != nil {
			userIDs = append(userIDs, *c.CreateBy)
		}
		if c.UpdateBy != nil {
			userIDs = append(userIDs, *c.UpdateBy)
		}
	}
	namesMap, err := database.GetUserNamesMap(r.db, userIDs)
	if err != nil {
		return
	}
	for _, c := range companies {
		if c.CreateBy != nil {
			c.CreateByName = namesMap[*c.CreateBy]
		}
		if c.UpdateBy != nil {
			c.UpdateByName = namesMap[*c.UpdateBy]
		}
	}
}

func (r *companyRepository) Create(company *domain.Company, req *domain.CreateCompanyRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create company
		if err := tx.Create(company).Error; err != nil {
			if strings.Contains(err.Error(), "idx_administrative_companies_name") || strings.Contains(err.Error(), "23505") {
				return fmt.Errorf("ya existe un negocio aliado registrado con el nombre '%s'", company.Name)
			}
			return err
		}

		// 2. If validator credentials provided, create validator user in the same transaction
		if req != nil && strings.TrimSpace(req.ValidatorEmail) != "" && strings.TrimSpace(req.ValidatorPassword) != "" {
			// Find role dynamically by code = 'business_validator'
			var role roleDomain.Role
			if err := tx.Where("code = ?", "business_validator").First(&role).Error; err != nil {
				return errors.New("no se encontró el rol de validador de negocio ('business_validator') en el sistema")
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.ValidatorPassword), bcrypt.DefaultCost)
			if err != nil {
				return errors.New("error al procesar la contraseña del validador")
			}

			firstName := strings.TrimSpace(req.ValidatorFirstName)
			if firstName == "" {
				firstName = "Validador"
			}
			lastName := strings.TrimSpace(req.ValidatorLastName)
			if lastName == "" {
				lastName = company.Name
			}

			user := &userDomain.User{
				Email:     strings.TrimSpace(req.ValidatorEmail),
				Password:  string(hashedPassword),
				FirstName: firstName,
				LastName:  lastName,
				IsActive:  true,
				RoleID:    &role.ID,
			}

			if err := tx.Create(user).Error; err != nil {
				if strings.Contains(err.Error(), "users_email_key") || strings.Contains(err.Error(), "23505") {
					return fmt.Errorf("el correo '%s' ya está registrado en la plataforma", user.Email)
				}
				return err
			}

			// Update empresa_id directly on the user (without user_companies)
			if err := tx.Exec("UPDATE administrative.users SET empresa_id = ? WHERE id = ?", company.ID, user.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *companyRepository) FindByID(id uint) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.First(&company, "id = ?", id).Error; err != nil {
		return nil, err
	}
	r.populateAudits([]*domain.Company{&company})
	return &company, nil
}

func (r *companyRepository) FindAll() ([]domain.Company, error) {
	var companies []domain.Company
	if err := r.db.Order("create_at DESC").Find(&companies).Error; err != nil {
		return nil, err
	}
	companiesPtrs := make([]*domain.Company, len(companies))
	for i := range companies {
		companiesPtrs[i] = &companies[i]
	}
	r.populateAudits(companiesPtrs)
	return companies, nil
}

func (r *companyRepository) Update(company *domain.Company) error {
	if err := r.db.Save(company).Error; err != nil {
		if strings.Contains(err.Error(), "idx_administrative_companies_name") || strings.Contains(err.Error(), "23505") {
			return fmt.Errorf("ya existe un negocio aliado registrado con el nombre '%s'", company.Name)
		}
		return err
	}
	return nil
}

func (r *companyRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Company{}, "id = ?", id).Error
}
