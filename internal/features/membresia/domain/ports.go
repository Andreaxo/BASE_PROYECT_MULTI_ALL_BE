package domain

// MembresiaRepository defines the secondary port for membership data persistence.
type MembresiaRepository interface {
	Create(membresia *Membresia) error
	FindByID(id uint) (*Membresia, error)
	FindByUsuarioID(usuarioID uint) (*Membresia, error)
	Update(membresia *Membresia) error
	FindActivasVencidas() ([]Membresia, error) // fecha_fin < now() AND estado = 'activa'
	FindAll() ([]Membresia, error)
}

// PagoMembresiaRepository defines the secondary port for payment data persistence.
type PagoMembresiaRepository interface {
	Create(pago *PagoMembresia) error
	FindByReferencia(referencia string) (*PagoMembresia, error)
	CountAprobadosByMembresia(membresiaID uint) (int64, error)
}

// MembresiaService defines the primary port for membership business operations.
type MembresiaService interface {
	IniciarPago(usuarioID uint, req *IniciarPagoRequest) (*IniciarPagoResponse, error)
	ProcessWebhook(body []byte, signature string) error
	GetMiMembresia(usuarioID uint) (*MembresiaResponse, error)
	CancelarRenovacion(usuarioID uint) (*MembresiaResponse, error)
	RenovarVencidas() error
	GetAllMembresias() ([]MembresiaResponse, error)
	SimularPago(usuarioID uint, status string) (*MembresiaResponse, error)
}
