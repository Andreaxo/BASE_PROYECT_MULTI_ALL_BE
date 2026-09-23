package access_control

import (
	"crypto/rand"
	"log"
	"math/big"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	companyDomain "multicliente-backend/internal/features/company/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
)

func generateCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "REF12345"
		}
		code[i] = charset[n.Int64()]
	}
	return string(code)
}

// SeedUsers seeds default administrator user and updates missing referral codes.
func SeedUsers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&userDomain.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminRoleID := uint(1)
		companies := []companyDomain.Company{
			{ID: 1, Name: "Empresa Base Demo", IsActive: true},
		}

		admin := &userDomain.User{
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
			RoleID:    &adminRoleID,
			Companies: companies,
			CodeRefer: "ADMIN123",
		}

		if err := db.Create(admin).Error; err != nil {
			return err
		}
		log.Println("✅ Admin user seeded successfully (admin@example.com / admin123)")
	}

	// Seed demo Negocio user for testing validation if not present
	var negocioCount int64
	db.Model(&userDomain.User{}).Where("email = ?", "negocio@example.com").Count(&negocioCount)
	if negocioCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("negocio123"), bcrypt.DefaultCost)
		if err == nil {
			negocioRoleID := uint(7)
			negocioUser := &userDomain.User{
				Email:     "negocio@example.com",
				Password:  string(hashedPassword),
				FirstName: "Validador",
				LastName:  "Negocio",
				IsActive:  true,
				RoleID:    &negocioRoleID,
				Companies: []companyDomain.Company{{ID: 1}},
				CodeRefer: "NEGOCIO1",
			}
			if err := db.Create(negocioUser).Error; err == nil {
				db.Exec("UPDATE administrative.users SET empresa_id = 1 WHERE id = ?", negocioUser.ID)
				log.Println("✅ Demo Negocio user seeded successfully (negocio@example.com / negocio123)")
			}
		}
	}

	// Seed demo Operador user for testing if not present
	var operadorCount int64
	db.Model(&userDomain.User{}).Where("email = ?", "operador@conexiate.com").Count(&operadorCount)
	if operadorCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Operador123*"), bcrypt.DefaultCost)
		if err == nil {
			operadorRoleID := uint(8)
			operadorUser := &userDomain.User{
				Email:     "operador@conexiate.com",
				Password:  string(hashedPassword),
				FirstName: "Operador",
				LastName:  "Atención",
				IsActive:  true,
				RoleID:    &operadorRoleID,
				Companies: []companyDomain.Company{{ID: 1}},
				CodeRefer: "OPERADOR",
			}
			if err := db.Create(operadorUser).Error; err == nil {
				db.Exec("UPDATE administrative.users SET empresa_id = 1 WHERE id = ?", operadorUser.ID)
				log.Println("✅ Demo Operador user seeded successfully (operador@conexiate.com / Operador123*)")
			}
		}
	}

	// Backfill missing code_refer and empresa_id for existing users
	var emptyUsers []userDomain.User
	if err := db.Where("code_refer IS NULL OR code_refer = ''").Find(&emptyUsers).Error; err == nil {
		for _, u := range emptyUsers {
			code := generateCode()
			db.Model(&userDomain.User{}).Where("id = ?", u.ID).Update("code_refer", code)
		}
	}

	// Backfill empresa_id = 1 for admin and demo users if null
	db.Exec("UPDATE administrative.users SET empresa_id = 1 WHERE empresa_id IS NULL OR empresa_id = 0")

	return nil
}
