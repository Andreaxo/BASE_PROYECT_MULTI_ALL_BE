package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"multicliente-backend/internal/features/benefit/domain"
)

type benefitRepository struct {
	db *gorm.DB
}

func NewBenefitRepository(db *gorm.DB) domain.BenefitRepository {
	return &benefitRepository{db: db}
}

func (r *benefitRepository) Create(benefit *domain.Benefit) error {
	return r.db.Create(benefit).Error
}

func (r *benefitRepository) FindByID(id uint) (*domain.Benefit, error) {
	var benefit domain.Benefit

	if err := r.db.First(&benefit, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &benefit, nil
}

func (r *benefitRepository) FindAll() ([]domain.Benefit, error) {
	benefits := make([]domain.Benefit, 0)

	type BenefitWithCompany struct {
		domain.Benefit
		SuscripcionEstado *string    `gorm:"column:suscripcion_estado"`
		FechaFinPrueba    *time.Time `gorm:"column:fecha_fin_prueba"`
	}

	var results []BenefitWithCompany
	if err := r.db.
		Table("administrative.benefit").
		Select("benefit.*, companies.suscripcion_estado, companies.fecha_fin_prueba").
		Joins("LEFT JOIN administrative.companies ON companies.id = benefit.company_benefit").
		Find(&results).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	for _, res := range results {
		b := res.Benefit
		isCompanyUnpaid := false
		if b.CompanyBenefits != nil && *b.CompanyBenefits > 0 {
			if res.SuscripcionEstado == nil {
				isCompanyUnpaid = true
			} else {
				estado := *res.SuscripcionEstado
				if estado == "vencida" || estado == "suspendida" {
					isCompanyUnpaid = true
				} else if estado == "prueba" && res.FechaFinPrueba != nil && res.FechaFinPrueba.Before(now) {
					isCompanyUnpaid = true
				} else if estado != "activa" && estado != "prueba" {
					isCompanyUnpaid = true
				}
			}
		}

		if isCompanyUnpaid {
			b.IsActive = false
			b.Estado = "deshabilitado_falta_pago"
		}

		benefits = append(benefits, b)
	}

	return benefits, nil
}

func (r *benefitRepository) FindAllByCompany(companyID uint) ([]domain.Benefit, error) {
	benefits := make([]domain.Benefit, 0)
	if err := r.db.Where("company_benefit = ?", companyID).Find(&benefits).Error; err != nil {
		return nil, err
	}

	var company struct {
		SuscripcionEstado string     `gorm:"column:suscripcion_estado"`
		FechaFinPrueba    *time.Time `gorm:"column:fecha_fin_prueba"`
	}
	if err := r.db.Table("administrative.companies").
		Select("suscripcion_estado, fecha_fin_prueba").
		Where("id = ?", companyID).
		Scan(&company).Error; err == nil {

		now := time.Now()
		isExpiredTrial := company.SuscripcionEstado == "prueba" &&
			company.FechaFinPrueba != nil &&
			company.FechaFinPrueba.Before(now)

		isNotPaid := company.SuscripcionEstado == "vencida" ||
			company.SuscripcionEstado == "suspendida" ||
			isExpiredTrial

		if isNotPaid {
			for i := range benefits {
				benefits[i].IsActive = false
				benefits[i].Estado = "deshabilitado_falta_pago"
			}
		}
	}

	return benefits, nil
}

func (r *benefitRepository) Update(benefit *domain.Benefit) error {
	return r.db.Save(benefit).Error
}

func (r *benefitRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Benefit{}, "id = ?", id).Error
}
