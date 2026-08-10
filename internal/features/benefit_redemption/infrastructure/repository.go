package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"multicliente-backend/internal/features/benefit_redemption/domain"
)

type redemptionRepository struct {
	db *gorm.DB
}

func NewRedemptionRepository(db *gorm.DB) domain.RedemptionRepository {
	return &redemptionRepository{db: db}
}

func (r *redemptionRepository) Create(redemption *domain.BenefitRedemption) error {
	return r.db.Create(redemption).Error
}

func (r *redemptionRepository) FindByCode(code string) (*domain.BenefitRedemption, error) {
	var redemption domain.BenefitRedemption
	if err := r.db.
		Preload("Benefit").
		First(&redemption, "codigo_validacion = ?", code).Error; err != nil {
		return nil, err
	}
	return &redemption, nil
}

func (r *redemptionRepository) FindByUsuarioID(usuarioID uint) ([]domain.BenefitRedemption, error) {
	var redemptions []domain.BenefitRedemption
	if err := r.db.
		Preload("Benefit").
		Where("usuario_id = ?", usuarioID).
		Order("fecha_redencion DESC").
		Find(&redemptions).Error; err != nil {
		return nil, err
	}
	return redemptions, nil
}

func (r *redemptionRepository) FindByEmpresaID(empresaID uint) ([]domain.BenefitRedemption, error) {
	var redemptions []domain.BenefitRedemption
	if err := r.db.
		Preload("Benefit").
		Joins("JOIN administrative.benefit b ON b.id = administrative.benefit_redemption.benefit_id").
		Where("b.company_benefit = ?", empresaID).
		Order("administrative.benefit_redemption.fecha_redencion DESC").
		Find(&redemptions).Error; err != nil {
		return nil, err
	}
	return redemptions, nil
}

func (r *redemptionRepository) FindAll() ([]domain.BenefitRedemption, error) {
	var redemptions []domain.BenefitRedemption
	if err := r.db.
		Preload("Benefit").
		Order("fecha_redencion DESC").
		Find(&redemptions).Error; err != nil {
		return nil, err
	}
	return redemptions, nil
}

// MarkAsUsed atomically updates the redemption to 'usado' only if it's currently 'generado'.
// Returns rows affected: 0 means the row was no longer in 'generado' state (race condition or already used).
func (r *redemptionRepository) MarkAsUsed(redemptionID uint, validadoPor uint) (int64, error) {
	now := time.Now()
	result := r.db.Model(&domain.BenefitRedemption{}).
		Where("id = ? AND estado = ?", redemptionID, domain.EstadoGenerado).
		Updates(map[string]interface{}{
			"estado":       domain.EstadoUsado,
			"fecha_uso":    now,
			"validado_por": validadoPor,
		})
	return result.RowsAffected, result.Error
}

// MarkAsExpired atomically updates the redemption to 'vencido' only if it's currently 'generado'.
// Uses the same atomic WHERE pattern as MarkAsUsed for concurrency safety.
func (r *redemptionRepository) MarkAsExpired(redemptionID uint) (int64, error) {
	result := r.db.Model(&domain.BenefitRedemption{}).
		Where("id = ? AND estado = ?", redemptionID, domain.EstadoGenerado).
		Update("estado", domain.EstadoVencido)
	return result.RowsAffected, result.Error
}

func (r *redemptionRepository) FindBenefitByID(id uint) (*domain.BenefitForRedemption, error) {
	var benefit domain.BenefitForRedemption
	if err := r.db.First(&benefit, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &benefit, nil
}

// GetUserEmpresaID retrieves empresa_id directly from the users table.
// This is necessary because empresa_id is not included in the JWT claims.
func (r *redemptionRepository) GetUserEmpresaID(userID uint) (uint, error) {
	var empresaID uint
	if err := r.db.Table("administrative.users").
		Select("empresa_id").
		Where("id = ?", userID).
		Scan(&empresaID).Error; err != nil {
		return 0, err
	}
	return empresaID, nil
}
