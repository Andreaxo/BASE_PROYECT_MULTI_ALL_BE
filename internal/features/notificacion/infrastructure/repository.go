package infrastructure

import (
	"gorm.io/gorm"

	"multicliente-backend/internal/features/notificacion/domain"
)

type notificacionRepository struct {
	db *gorm.DB
}

// NewNotificacionRepository crea una nueva instancia del repositorio de notificaciones.
func NewNotificacionRepository(db *gorm.DB) domain.NotificacionRepository {
	return &notificacionRepository{db: db}
}

func (r *notificacionRepository) FindByUsuarioID(usuarioID uint, limit, offset int) ([]domain.Notificacion, error) {
	var notificaciones []domain.Notificacion
	query := r.db.Where("usuario_id = ?", usuarioID).Order("create_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&notificaciones).Error
	return notificaciones, err
}

func (r *notificacionRepository) CountNoLeidas(usuarioID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notificacion{}).
		Where("usuario_id = ? AND leido = false", usuarioID).
		Count(&count).Error
	return count, err
}

func (r *notificacionRepository) MarcarLeida(id uint, usuarioID uint) error {
	return r.db.Model(&domain.Notificacion{}).
		Where("id = ? AND usuario_id = ?", id, usuarioID).
		Update("leido", true).Error
}

func (r *notificacionRepository) MarcarTodasLeidas(usuarioID uint) error {
	return r.db.Model(&domain.Notificacion{}).
		Where("usuario_id = ? AND leido = false", usuarioID).
		Update("leido", true).Error
}

func (r *notificacionRepository) Create(notificacion *domain.Notificacion) error {
	return r.db.Create(notificacion).Error
}

func (r *notificacionRepository) CreateBatch(notificaciones []domain.Notificacion) error {
	if len(notificaciones) == 0 {
		return nil
	}
	return r.db.Create(&notificaciones).Error
}
