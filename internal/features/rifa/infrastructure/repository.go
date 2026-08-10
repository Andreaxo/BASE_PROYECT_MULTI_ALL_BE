package infrastructure

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"multicliente-backend/internal/features/rifa/domain"
)

// --- RifaRepository ---

type rifaRepository struct {
	db *gorm.DB
}

func NewRifaRepository(db *gorm.DB) domain.RifaRepository {
	return &rifaRepository{db: db}
}

func (r *rifaRepository) Create(rifa *domain.Rifa) error {
	return r.db.Create(rifa).Error
}

func (r *rifaRepository) FindByID(id uint) (*domain.Rifa, error) {
	var rifa domain.Rifa
	if err := r.db.First(&rifa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rifa, nil
}

// FindByIDForUpdate locks the rifa row within an existing transaction.
// Used to serialize numero_participacion generation per rifa.
func (r *rifaRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*domain.Rifa, error) {
	var rifa domain.Rifa
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&rifa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rifa, nil
}

func (r *rifaRepository) FindActiva() (*domain.Rifa, error) {
	var rifa domain.Rifa
	if err := r.db.First(&rifa, "estado = ?", domain.EstadoActiva).Error; err != nil {
		return nil, err
	}
	return &rifa, nil
}

func (r *rifaRepository) FindAll() ([]domain.Rifa, error) {
	var rifas []domain.Rifa
	if err := r.db.Order("create_at DESC").Find(&rifas).Error; err != nil {
		return nil, err
	}
	return rifas, nil
}

func (r *rifaRepository) Update(rifa *domain.Rifa) error {
	return r.db.Save(rifa).Error
}

func (r *rifaRepository) UpdateTx(tx *gorm.DB, rifa *domain.Rifa) error {
	return tx.Save(rifa).Error
}

// --- ParticipacionRepository ---

type participacionRepository struct {
	db *gorm.DB
}

func NewParticipacionRepository(db *gorm.DB) domain.ParticipacionRepository {
	return &participacionRepository{db: db}
}

func (r *participacionRepository) CreateTx(tx *gorm.DB, p *domain.ParticipacionRifa) error {
	return tx.Create(p).Error
}

func (r *participacionRepository) CountByRifaTx(tx *gorm.DB, rifaID uint) (int64, error) {
	var count int64
	if err := tx.Model(&domain.ParticipacionRifa{}).
		Where("rifa_id = ?", rifaID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *participacionRepository) FindByRifaID(rifaID uint) ([]domain.ParticipacionRifa, error) {
	var participaciones []domain.ParticipacionRifa
	if err := r.db.
		Preload("Usuario").
		Where("rifa_id = ?", rifaID).
		Order("numero_participacion ASC").
		Find(&participaciones).Error; err != nil {
		return nil, err
	}
	return participaciones, nil
}

func (r *participacionRepository) FindByUsuarioAndRifa(usuarioID uint, rifaID uint) ([]domain.ParticipacionRifa, error) {
	var participaciones []domain.ParticipacionRifa
	if err := r.db.
		Preload("Usuario").
		Where("usuario_id = ? AND rifa_id = ?", usuarioID, rifaID).
		Order("numero_participacion ASC").
		Find(&participaciones).Error; err != nil {
		return nil, err
	}
	return participaciones, nil
}

func (r *participacionRepository) FindByID(id uint) (*domain.ParticipacionRifa, error) {
	var p domain.ParticipacionRifa
	if err := r.db.
		Preload("Usuario").
		First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// --- GanadorRepository ---

type ganadorRepository struct {
	db *gorm.DB
}

func NewGanadorRepository(db *gorm.DB) domain.GanadorRepository {
	return &ganadorRepository{db: db}
}

func (r *ganadorRepository) Create(g *domain.GanadorRifa) error {
	return r.db.Create(g).Error
}

func (r *ganadorRepository) FindByRifaID(rifaID uint) ([]domain.GanadorRifa, error) {
	var ganadores []domain.GanadorRifa
	if err := r.db.
		Preload("Participacion").
		Preload("Participacion.Usuario").
		Where("rifa_id = ?", rifaID).
		Find(&ganadores).Error; err != nil {
		return nil, err
	}
	return ganadores, nil
}

func (r *ganadorRepository) FindByID(id uint) (*domain.GanadorRifa, error) {
	var g domain.GanadorRifa
	if err := r.db.
		Preload("Participacion").
		Preload("Participacion.Usuario").
		First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *ganadorRepository) Update(g *domain.GanadorRifa) error {
	return r.db.Save(g).Error
}
