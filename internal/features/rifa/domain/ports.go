package domain

import "gorm.io/gorm"

// RifaRepository defines the secondary port for rifa data persistence.
type RifaRepository interface {
	Create(rifa *Rifa) error
	FindByID(id uint) (*Rifa, error)
	FindByIDForUpdate(tx *gorm.DB, id uint) (*Rifa, error)
	FindActiva() (*Rifa, error)
	FindAll() ([]Rifa, error)
	Update(rifa *Rifa) error
	UpdateTx(tx *gorm.DB, rifa *Rifa) error
}

// ParticipacionRepository defines the secondary port for participacion_rifa data persistence.
type ParticipacionRepository interface {
	CreateTx(tx *gorm.DB, p *ParticipacionRifa) error
	CountByRifaTx(tx *gorm.DB, rifaID uint) (int64, error)
	FindByRifaID(rifaID uint) ([]ParticipacionRifa, error)
	FindByUsuarioAndRifa(usuarioID uint, rifaID uint) ([]ParticipacionRifa, error)
	FindByID(id uint) (*ParticipacionRifa, error)
}

// GanadorRepository defines the secondary port for ganador_rifa data persistence.
type GanadorRepository interface {
	Create(g *GanadorRifa) error
	FindByRifaID(rifaID uint) ([]GanadorRifa, error)
	FindByID(id uint) (*GanadorRifa, error)
	Update(g *GanadorRifa) error
}

// RifaService defines the primary port for raffle business operations.
type RifaService interface {
	CreateRifa(req *CreateRifaRequest, createdBy *uint) (*RifaResponse, error)
	UpdateRifa(id uint, req *UpdateRifaRequest, updatedBy *uint) (*RifaResponse, error)
	ActivarRifa(id uint, updatedBy *uint) (*RifaResponse, error)
	CerrarRifa(id uint, updatedBy *uint) (*RifaResponse, error)
	SortearRifa(id uint, updatedBy *uint) (*RifaResponse, error)
	CancelarRifa(id uint, updatedBy *uint) (*RifaResponse, error)
	GetRifaActiva() (*RifaResponse, error)
	GetRifaByID(id uint) (*RifaResponse, error)
	GetAllRifas() ([]*RifaResponse, error)
	AgregarParticipacion(rifaID uint, usuarioID uint, origen string) (*ParticipacionResponse, error)
	GetParticipacionesByRifa(rifaID uint) ([]*ParticipacionResponse, error)
	GetMisParticipaciones(usuarioID uint) ([]*ParticipacionResponse, error)
	RegistrarGanador(rifaID uint, participacionID uint) (*GanadorResponse, error)
	ActualizarEstadoEntrega(ganadorID uint, nuevoEstado string) (*GanadorResponse, error)
	GetGanadoresByRifa(rifaID uint) ([]*GanadorResponse, error)
	GetHistorial() ([]*RifaHistorialResponse, error)

	// OtorgarRecompensaReferido creates a raffle participation for a referente
	// when one of their referrals becomes 'afiliado'. Returns nil if successful,
	// error if no active rifa exists or the operation fails.
	OtorgarRecompensaReferido(referenteID uint) error

	// OtorgarRecompensasPendientes finds all referidos with estado='afiliado' AND
	// recompensa_otorgada=false, and grants them raffle participations.
	// Called when activating a new rifa.
	OtorgarRecompensasPendientes(rifaID uint) error
}
