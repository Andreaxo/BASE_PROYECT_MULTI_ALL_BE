package infrastructure

import (
	"gorm.io/gorm"

	"multicliente-backend/internal/features/membresia/domain"
)

type membresiaRepository struct {
	db *gorm.DB
}

func NewMembresiaRepository(db *gorm.DB) domain.MembresiaRepository {
	return &membresiaRepository{db: db}
}

func (r *membresiaRepository) Create(membresia *domain.Membresia) error {
	return r.db.Create(membresia).Error
}

func (r *membresiaRepository) FindByID(id uint) (*domain.Membresia, error) {
	var m domain.Membresia
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *membresiaRepository) FindByUsuarioID(usuarioID uint) (*domain.Membresia, error) {
	var m domain.Membresia
	if err := r.db.First(&m, "usuario_id = ?", usuarioID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *membresiaRepository) Update(membresia *domain.Membresia) error {
	return r.db.Save(membresia).Error
}

// FindActivasVencidas finds memberships that are active but past their end date.
func (r *membresiaRepository) FindActivasVencidas() ([]domain.Membresia, error) {
	var membresias []domain.Membresia
	if err := r.db.
		Where("estado = ? AND fecha_fin < NOW() AND renovacion_automatica = true", domain.MembresiaActiva).
		Find(&membresias).Error; err != nil {
		return nil, err
	}
	return membresias, nil
}

func (r *membresiaRepository) FindAll() ([]domain.Membresia, error) {
	var membresias []domain.Membresia
	if err := r.db.Order("create_at DESC").Find(&membresias).Error; err != nil {
		return nil, err
	}
	return membresias, nil
}

// --- PagoMembresia Repository ---

type pagoMembresiaRepository struct {
	db *gorm.DB
}

func NewPagoMembresiaRepository(db *gorm.DB) domain.PagoMembresiaRepository {
	return &pagoMembresiaRepository{db: db}
}

func (r *pagoMembresiaRepository) Create(pago *domain.PagoMembresia) error {
	return r.db.Create(pago).Error
}

func (r *pagoMembresiaRepository) FindByReferencia(referencia string) (*domain.PagoMembresia, error) {
	var pago domain.PagoMembresia
	if err := r.db.First(&pago, "referencia_pasarela = ?", referencia).Error; err != nil {
		return nil, err
	}
	return &pago, nil
}

func (r *pagoMembresiaRepository) CountAprobadosByMembresia(membresiaID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.PagoMembresia{}).
		Where("membresia_id = ? AND estado = ?", membresiaID, domain.PagoEstadoAprobado).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
