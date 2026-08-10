package seeds

import (
	"log"

	"gorm.io/gorm"

	benefitDomain "multicliente-backend/internal/features/benefit/domain"
)

// SeedBenefits populates initial demo benefits into the database.
func SeedBenefits(db *gorm.DB) error {
	var count int64
	if err := db.Model(&benefitDomain.Benefit{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		companyID := uint(1)
		benefits := []benefitDomain.Benefit{
			{
				ID:              1,
				Name:            "Bono de Alimentación",
				Description:     "Auxilio mensual de alimentación para empleados",
				CompanyBenefits: &companyID,
			},
			{
				ID:              2,
				Name:            "Seguro Médico Complementario",
				Description:     "Cobertura médica privada del 80%",
				CompanyBenefits: &companyID,
			},
			{
				ID:              3,
				Name:            "Bono de Transporte",
				Description:     "Subsidio mensual para desplazamientos",
				CompanyBenefits: &companyID,
			},
			{
				ID:              4,
				Name:            "Descuento en Gimnasio",
				Description:     "Pase libre con 50% de descuento en cadenas afiliadas",
				CompanyBenefits: &companyID,
			},
		}

		for _, b := range benefits {
			if err := db.Create(&b).Error; err != nil {
				log.Printf("⚠️ Warning: could not seed benefit %s: %v", b.Name, err)
			}
		}
		log.Println("✅ Benefits seeded successfully")
	}

	return nil
}
