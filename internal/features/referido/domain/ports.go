package domain

// ReferidoRepository defines the secondary port for referral data persistence.
type ReferidoRepository interface {
	Create(referido *Referido) error
	FindByID(id uint) (*Referido, error)
	FindByReferenteID(referenteID uint) ([]Referido, error)
	FindByEmail(email string) (*Referido, error)
	FindPendingByEmail(email string) (*Referido, error)
	FindRegistradoByUsuarioID(usuarioReferidoID uint) (*Referido, error)
	FindAll() ([]Referido, error)
	Update(referido *Referido) error
	Delete(id uint) error
	FindPendingRewards() ([]Referido, error)
}

// ReferidoService defines the primary port for referral business operations.
type ReferidoService interface {
	CreateReferido(req *CreateReferidoRequest, referenteID uint) (*ReferidoResponse, error)
	GetReferidoByID(id uint) (*ReferidoResponse, error)
	GetReferidosByReferente(referenteID uint) ([]*ReferidoResponse, error)
	GetAllReferidos() ([]*ReferidoResponse, error)
	ConvertirReferido(email string, usuarioReferidoID uint) (*ReferidoResponse, error)
	AfiliarReferido(id uint) (*ReferidoResponse, error)
	OtorgarRecompensa(id uint) (*ReferidoResponse, error)
	DeleteReferido(id uint) error
}
