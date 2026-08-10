package infrastructure

import (
	"gorm.io/gorm"

	"multicliente-backend/internal/features/referido/domain"
)

type referidoRepository struct {
	db *gorm.DB
}

// NewReferidoRepository creates a new GORM-based ReferidoRepository.
func NewReferidoRepository(db *gorm.DB) domain.ReferidoRepository {
	return &referidoRepository{db: db}
}

func (r *referidoRepository) Create(referido *domain.Referido) error {
	return r.db.Create(referido).Error
}

func (r *referidoRepository) FindByID(id uint) (*domain.Referido, error) {
	var referido domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		First(&referido, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &referido, nil
}

func (r *referidoRepository) FindByReferenteID(referenteID uint) ([]domain.Referido, error) {
	var referidos []domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		Where("usuario_referente_id = ?", referenteID).
		Order("fecha_referido DESC").
		Find(&referidos).Error; err != nil {
		return nil, err
	}
	return referidos, nil
}

func (r *referidoRepository) FindByEmail(email string) (*domain.Referido, error) {
	var referido domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		First(&referido, "email_referido = ?", email).Error; err != nil {
		return nil, err
	}
	return &referido, nil
}

func (r *referidoRepository) FindPendingByEmail(email string) (*domain.Referido, error) {
	var referido domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		First(&referido, "email_referido = ? AND estado = ?", email, domain.EstadoPendiente).Error; err != nil {
		return nil, err
	}
	return &referido, nil
}

func (r *referidoRepository) FindAll() ([]domain.Referido, error) {
	var referidos []domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		Order("fecha_referido DESC").
		Find(&referidos).Error; err != nil {
		return nil, err
	}
	return referidos, nil
}

func (r *referidoRepository) Update(referido *domain.Referido) error {
	return r.db.Save(referido).Error
}

func (r *referidoRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Referido{}, "id = ?", id).Error
}

func (r *referidoRepository) FindPendingRewards() ([]domain.Referido, error) {
	var referidos []domain.Referido
	if err := r.db.
		Preload("Referente").
		Preload("Referido_").
		Where("estado = ? AND recompensa_otorgada = ?", domain.EstadoAfiliado, false).
		Order("fecha_referido ASC").
		Find(&referidos).Error; err != nil {
		return nil, err
	}
	return referidos, nil
}
