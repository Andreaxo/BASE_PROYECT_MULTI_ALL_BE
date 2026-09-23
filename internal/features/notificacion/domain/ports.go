package domain

// NotificacionRepository define los métodos de acceso a datos para notificaciones.
type NotificacionRepository interface {
	FindByUsuarioID(usuarioID uint, limit, offset int) ([]Notificacion, error)
	CountNoLeidas(usuarioID uint) (int64, error)
	MarcarLeida(id uint, usuarioID uint) error
	MarcarTodasLeidas(usuarioID uint) error
	Create(notificacion *Notificacion) error
	CreateBatch(notificaciones []Notificacion) error
}

// NotificacionService define los casos de uso para la gestión y emisión de notificaciones.
type NotificacionService interface {
	GetNotificaciones(usuarioID uint, limit, offset int) ([]Notificacion, error)
	GetCountNoLeidas(usuarioID uint) (int64, error)
	MarcarLeida(id uint, usuarioID uint) error
	MarcarTodasLeidas(usuarioID uint) error
	NotificarPagoSuperadmins(titulo, mensaje, tipo string, pagoID uint, referenciaTipo string) error
}
