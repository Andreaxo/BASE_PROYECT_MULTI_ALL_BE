package infrastructure

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"multicliente-backend/internal/features/company/domain"
	membresiaDomain "multicliente-backend/internal/features/membresia/domain"
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
		// Generate codigo_empresa
		codigo, err := generateCodigoEmpresa()
		if err != nil {
			return errors.New("error al generar código de empresa")
		}
		company.CodigoEmpresa = &codigo

		// Set subscription state and trial period
		company.SuscripcionEstado = "prueba"
		mesesStr := membresiaDomain.GetConfigValue(tx, "meses_gratis_negocio", "12")
		meses, err := strconv.Atoi(mesesStr)
		if err != nil || meses <= 0 {
			meses = 12
		}
		finPrueba := time.Now().AddDate(0, meses, 0)
		company.FechaFinPrueba = &finPrueba

		// Validate uniqueness of name, nit, and razon_social proactively
		var count int64
		if err := tx.Model(&domain.Company{}).Where("LOWER(TRIM(name)) = LOWER(TRIM(?))", company.Name).Count(&count).Error; err == nil && count > 0 {
			return errors.New("Ya existe una empresa registrada con este nombre comercial.")
		}
		if company.NIT != nil && *company.NIT > 0 {
			if err := tx.Model(&domain.Company{}).Where("nit = ?", *company.NIT).Count(&count).Error; err == nil && count > 0 {
				return errors.New("Ya existe una empresa registrada con este NIT.")
			}
		}
		if strings.TrimSpace(company.RazonSocial) != "" {
			if err := tx.Model(&domain.Company{}).Where("LOWER(TRIM(razon_social)) = LOWER(TRIM(?))", company.RazonSocial).Count(&count).Error; err == nil && count > 0 {
				return errors.New("Ya existe una empresa registrada con esta razón social.")
			}
		}

		// 1. Create company
		if err := tx.Create(company).Error; err != nil {
			return parseCompanyDBError(err, company)
		}

		// 2. If validator credentials provided, create validator user in the same transaction
		if req != nil && strings.TrimSpace(req.ValidatorEmail) != "" && strings.TrimSpace(req.ValidatorPassword) != "" {
			valEmail := strings.TrimSpace(req.ValidatorEmail)
			var userCount int64
			if err := tx.Model(&userDomain.User{}).Where("LOWER(TRIM(email)) = LOWER(TRIM(?))", valEmail).Count(&userCount).Error; err == nil && userCount > 0 {
				return errors.New("El correo electrónico del validador ya se encuentra registrado en la plataforma.")
			}

			// Find role dynamically by code = 'business_validator'
			var role roleDomain.Role
			if err := tx.Where("code = ?", "business_validator").First(&role).Error; err != nil {
				return errors.New("No se encontró el rol de validador de negocio ('business_validator') en el sistema.")
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.ValidatorPassword), bcrypt.DefaultCost)
			if err != nil {
				return errors.New("Error al procesar la contraseña del validador.")
			}

			firstName := strings.TrimSpace(req.ValidatorFirstName)
			if firstName == "" {
				firstName = "Validador"
			}
			lastName := strings.TrimSpace(req.ValidatorLastName)
			if lastName == "" {
				lastName = company.Name
			}

			codeRefer, err := generateCodigoEmpresa()
			if err != nil {
				return errors.New("Error al generar código de referencia para el validador.")
			}

			user := &userDomain.User{
				Email:     valEmail,
				Password:  string(hashedPassword),
				FirstName: firstName,
				LastName:  lastName,
				IsActive:  true,
				RoleID:    &role.ID,
				CodeRefer: codeRefer,
			}

			if err := tx.Create(user).Error; err != nil {
				if strings.Contains(err.Error(), "users_email_key") || strings.Contains(err.Error(), "23505") {
					return fmt.Errorf("El correo electrónico '%s' ya se encuentra registrado en la plataforma.", user.Email)
				}
				return err
			}

			// Associate validator user with the company in user_companies and update empresa_id
			if err := tx.Exec("UPDATE administrative.users SET empresa_id = ? WHERE id = ?", company.ID, user.ID).Error; err != nil {
				return err
			}
			if err := tx.Exec("INSERT INTO administrative.user_companies (user_id, company_id) VALUES (?, ?) ON CONFLICT DO NOTHING", user.ID, company.ID).Error; err != nil {
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
	var count int64
	if err := r.db.Model(&domain.Company{}).Where("LOWER(TRIM(name)) = LOWER(TRIM(?)) AND id != ?", company.Name, company.ID).Count(&count).Error; err == nil && count > 0 {
		return errors.New("Ya existe una empresa registrada con este nombre comercial.")
	}
	if company.NIT != nil && *company.NIT > 0 {
		if err := r.db.Model(&domain.Company{}).Where("nit = ? AND id != ?", *company.NIT, company.ID).Count(&count).Error; err == nil && count > 0 {
			return errors.New("Ya existe una empresa registrada con este NIT.")
		}
	}
	if strings.TrimSpace(company.RazonSocial) != "" {
		if err := r.db.Model(&domain.Company{}).Where("LOWER(TRIM(razon_social)) = LOWER(TRIM(?)) AND id != ?", company.RazonSocial, company.ID).Count(&count).Error; err == nil && count > 0 {
			return errors.New("Ya existe una empresa registrada con esta razón social.")
		}
	}

	if err := r.db.Save(company).Error; err != nil {
		return parseCompanyDBError(err, company)
	}
	return nil
}

func parseCompanyDBError(err error, company *domain.Company) error {
	if err == nil {
		return nil
	}
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "nit") && (strings.Contains(errStr, "uni_") || strings.Contains(errStr, "idx_") || strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique")) {
		return errors.New("Ya existe una empresa registrada con este NIT.")
	}
	if strings.Contains(errStr, "razon_social") && (strings.Contains(errStr, "uni_") || strings.Contains(errStr, "idx_") || strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique")) {
		return errors.New("Ya existe una empresa registrada con esta razón social.")
	}
	if strings.Contains(errStr, "name") && (strings.Contains(errStr, "companies") || strings.Contains(errStr, "23505") || strings.Contains(errStr, "unique")) {
		return errors.New("Ya existe una empresa registrada con este nombre comercial.")
	}
	if strings.Contains(errStr, "codigo_empresa") {
		return errors.New("Ya existe una empresa registrada con este código de empresa.")
	}
	if strings.Contains(errStr, "23505") {
		return errors.New("Ya existe un registro con estos datos en el sistema.")
	}
	return err
}

func (r *companyRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Block deletion if company has active benefits
		var activeBenefitsCount int64
		if err := tx.Table("administrative.benefit").
			Where("company_benefit = ? AND is_active = true AND estado = 'activo'", id).
			Count(&activeBenefitsCount).Error; err == nil && activeBenefitsCount > 0 {
			return errors.New("company_has_active_benefits")
		}

		// 2. Block deletion if company has active validators or active users assigned
		var activeUsersCount int64
		if err := tx.Table("administrative.users u").
			Where("u.is_active = true AND (u.empresa_id = ? OR u.id IN (SELECT user_id FROM administrative.user_companies WHERE company_id = ?))", id, id).
			Count(&activeUsersCount).Error; err == nil && activeUsersCount > 0 {
			return errors.New("company_has_active_users")
		}

		// 3. Delete inactive associations in user_companies
		if err := tx.Exec("DELETE FROM administrative.user_companies WHERE company_id = ?", id).Error; err != nil {
			return err
		}

		// 4. Unlink any remaining inactive users pointing to this company
		if err := tx.Exec("UPDATE administrative.users SET empresa_id = NULL WHERE empresa_id = ?", id).Error; err != nil {
			return err
		}

		// 5. Unlink/deactivate any remaining inactive benefits
		_ = tx.Exec("UPDATE administrative.benefit SET company_benefit = NULL, is_active = false WHERE company_benefit = ?", id)

		// 6. Delete company's inventory categories and articles if any exist
		_ = tx.Exec("DELETE FROM app.articles WHERE company_id = ?", id)
		_ = tx.Exec("DELETE FROM app.categories WHERE company_id = ?", id)

		// 7. Delete the company record
		res := tx.Delete(&domain.Company{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("empresa no encontrada")
		}
		return nil
	})
}

func (r *companyRepository) FindByCodigoEmpresa(codigo string) (*domain.Company, error) {
	var company domain.Company
	if err := r.db.First(&company, "codigo_empresa = ?", codigo).Error; err != nil {
		return nil, err
	}
	r.populateAudits([]*domain.Company{&company})
	return &company, nil
}

// generateCodigoEmpresa generates a random 8-character code without ambiguous characters.
func generateCodigoEmpresa() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No 0/O, 1/I
	const length = 8
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}
