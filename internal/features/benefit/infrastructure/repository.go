package infrastructure

import (
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

	if err := r.db.Find(&benefits).Error; err != nil {
		return nil, err
	}

	return benefits, nil
}

func (r *benefitRepository) FindAllByCompany(companyID uint) ([]domain.Benefit, error) {
	benefits := make([]domain.Benefit, 0)
	if err := r.db.Where("company_benefit = ?", companyID).Find(&benefits).Error; err != nil {
		return nil, err
	}
	return benefits, nil
}

func (r *benefitRepository) Update(benefit *domain.Benefit) error {
	return r.db.Save(benefit).Error
}

func (r *benefitRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Benefit{}, "id = ?", id).Error
}
