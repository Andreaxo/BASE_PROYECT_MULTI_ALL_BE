package seeds

import (
	"log"
	companyDomain "multicliente-backend/internal/features/company/domain"

	"gorm.io/gorm"
)

// SeedCompanies seeds the default test company.
func SeedCompanies(db *gorm.DB) error {
	companies := []companyDomain.Company{
		{ID: 1, Name: "Empresa Base Demo", IsActive: true},
	}
	for _, c := range companies {
		if err := db.FirstOrCreate(&c, companyDomain.Company{ID: c.ID}).Error; err != nil {
			return err
		}
	}
	log.Println("✅ Default company seeded")
	return nil
}
