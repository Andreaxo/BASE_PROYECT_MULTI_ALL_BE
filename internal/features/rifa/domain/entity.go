package domain

import (
	"time"

	userDomain "multicliente-backend/internal/features/user/domain"
)

// --- Estado constants for rifa ---
const (
	EstadoActiva    = "activa"
	EstadoCerrada   = "cerrada"
	EstadoSorteada  = "sorteada"
	EstadoCancelada = "cancelada"
)

// --- Origen constants for participacion_rifa ---
const (
	OrigenAfiliacion = "afiliacion"
	OrigenReferido   = "referido"
	OrigenBeneficio  = "beneficio"
	OrigenManual     = "manual"
)

// --- Estado entrega constants for ganador_rifa ---
const (
	EntregaPendiente  = "pendiente"
	EntregaNotificado = "notificado"
	EntregaEntregado  = "entregado"
)

// Rifa represents the administrative.rifa table.
type Rifa struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Nombre      string     `gorm:"type:varchar(150);not null" json:"nombre"`
	Descripcion *string    `gorm:"type:varchar(500)" json:"descripcion"`
	Premio      string     `gorm:"type:varchar(200);not null" json:"premio"`
	ImagenURL   *string    `gorm:"type:varchar(255)" json:"imagen_url"`
	FechaInicio time.Time  `gorm:"type:date;not null" json:"fecha_inicio"`
	FechaFin    time.Time  `gorm:"type:date;not null" json:"fecha_fin"`
	FechaSorteo time.Time  `gorm:"type:date;not null" json:"fecha_sorteo"`
	Estado      string     `gorm:"type:varchar(20);not null;default:'activa'" json:"estado"`
	CreateBy    *uint      `json:"create_by"`
	CreateAt    time.Time  `gorm:"type:timestamptz;not null;autoCreateTime" json:"create_at"`
	UpdateBy    *uint      `json:"update_by"`
	UpdateAt    *time.Time `gorm:"type:timestamptz" json:"update_at"`
}

func (Rifa) TableName() string {
	return "administrative.rifa"
}

// ParticipacionRifa represents the administrative.participacion_rifa table.
type ParticipacionRifa struct {
	ID                  uint             `gorm:"primaryKey" json:"id"`
	RifaID              uint             `gorm:"not null;index:idx_participacion_rifa" json:"rifa_id"`
	UsuarioID           uint             `gorm:"not null;index:idx_participacion_usuario" json:"usuario_id"`
	NumeroParticipacion string           `gorm:"type:varchar(20);not null" json:"numero_participacion"`
	Origen              string           `gorm:"type:varchar(20);not null" json:"origen"`
	FechaParticipacion  time.Time        `gorm:"type:timestamptz;not null;autoCreateTime" json:"fecha_participacion"`
	Rifa                *Rifa            `gorm:"foreignKey:RifaID" json:"rifa,omitempty"`
	Usuario             *userDomain.User `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
}

func (ParticipacionRifa) TableName() string {
	return "administrative.participacion_rifa"
}

// GanadorRifa represents the administrative.ganador_rifa table.
// The winning user is always obtained via JOIN:
// ganador_rifa.participacion_id → participacion_rifa.usuario_id → users.
type GanadorRifa struct {
	ID                uint               `gorm:"primaryKey" json:"id"`
	RifaID            uint               `gorm:"not null;index:idx_ganador_rifa" json:"rifa_id"`
	ParticipacionID   uint               `gorm:"not null;uniqueIndex" json:"participacion_id"`
	FechaNotificacion *time.Time         `gorm:"type:timestamptz" json:"fecha_notificacion"`
	EstadoEntrega     string             `gorm:"type:varchar(20);not null;default:'pendiente'" json:"estado_entrega"`
	Rifa              *Rifa              `gorm:"foreignKey:RifaID" json:"rifa,omitempty"`
	Participacion     *ParticipacionRifa `gorm:"foreignKey:ParticipacionID" json:"participacion,omitempty"`
}

func (GanadorRifa) TableName() string {
	return "administrative.ganador_rifa"
}

// --- Request DTOs ---

type CreateRifaRequest struct {
	Nombre      string  `json:"nombre" binding:"required"`
	Descripcion *string `json:"descripcion"`
	Premio      string  `json:"premio" binding:"required"`
	ImagenURL   *string `json:"imagen_url"`
	FechaInicio string  `json:"fecha_inicio" binding:"required"`
	FechaFin    string  `json:"fecha_fin" binding:"required"`
	FechaSorteo string  `json:"fecha_sorteo" binding:"required"`
}

type UpdateRifaRequest struct {
	Nombre      *string `json:"nombre"`
	Descripcion *string `json:"descripcion"`
	Premio      *string `json:"premio"`
	ImagenURL   *string `json:"imagen_url"`
	FechaInicio *string `json:"fecha_inicio"`
	FechaFin    *string `json:"fecha_fin"`
	FechaSorteo *string `json:"fecha_sorteo"`
}

type AgregarParticipacionManualRequest struct {
	UsuarioID uint `json:"usuario_id" binding:"required"`
}

type RegistrarGanadorRequest struct {
	ParticipacionID uint `json:"participacion_id" binding:"required"`
}

type ActualizarEntregaRequest struct {
	EstadoEntrega string `json:"estado_entrega" binding:"required"`
}

// --- Response DTOs ---

type RifaResponse struct {
	ID          uint       `json:"id"`
	Nombre      string     `json:"nombre"`
	Descripcion *string    `json:"descripcion"`
	Premio      string     `json:"premio"`
	ImagenURL   *string    `json:"imagen_url"`
	FechaInicio time.Time  `json:"fecha_inicio"`
	FechaFin    time.Time  `json:"fecha_fin"`
	FechaSorteo time.Time  `json:"fecha_sorteo"`
	Estado      string     `json:"estado"`
	CreateBy    *uint      `json:"create_by"`
	CreateAt    time.Time  `json:"create_at"`
	UpdateBy    *uint      `json:"update_by"`
	UpdateAt    *time.Time `json:"update_at"`
}

type ParticipacionResponse struct {
	ID                  uint      `json:"id"`
	RifaID              uint      `json:"rifa_id"`
	UsuarioID           uint      `json:"usuario_id"`
	NombreUsuario       string    `json:"nombre_usuario"`
	NumeroParticipacion string    `json:"numero_participacion"`
	Origen              string    `json:"origen"`
	FechaParticipacion  time.Time `json:"fecha_participacion"`
}

type GanadorResponse struct {
	ID                  uint       `json:"id"`
	RifaID              uint       `json:"rifa_id"`
	ParticipacionID     uint       `json:"participacion_id"`
	NumeroParticipacion string     `json:"numero_participacion"`
	UsuarioID           uint       `json:"usuario_id"`
	NombreGanador       string     `json:"nombre_ganador"`
	FechaNotificacion   *time.Time `json:"fecha_notificacion"`
	EstadoEntrega       string     `json:"estado_entrega"`
}

type RifaHistorialResponse struct {
	Rifa      RifaResponse      `json:"rifa"`
	Ganadores []GanadorResponse `json:"ganadores"`
}

// --- Converter functions ---

func ToRifaResponse(r *Rifa) *RifaResponse {
	return &RifaResponse{
		ID:          r.ID,
		Nombre:      r.Nombre,
		Descripcion: r.Descripcion,
		Premio:      r.Premio,
		ImagenURL:   r.ImagenURL,
		FechaInicio: r.FechaInicio,
		FechaFin:    r.FechaFin,
		FechaSorteo: r.FechaSorteo,
		Estado:      r.Estado,
		CreateBy:    r.CreateBy,
		CreateAt:    r.CreateAt,
		UpdateBy:    r.UpdateBy,
		UpdateAt:    r.UpdateAt,
	}
}

func ToRifaResponses(rifas []Rifa) []*RifaResponse {
	responses := make([]*RifaResponse, len(rifas))
	for i, r := range rifas {
		responses[i] = ToRifaResponse(&r)
	}
	return responses
}

func ToParticipacionResponse(p *ParticipacionRifa) *ParticipacionResponse {
	nombreUsuario := ""
	if p.Usuario != nil {
		nombreUsuario = p.Usuario.FirstName + " " + p.Usuario.LastName
	}
	return &ParticipacionResponse{
		ID:                  p.ID,
		RifaID:              p.RifaID,
		UsuarioID:           p.UsuarioID,
		NombreUsuario:       nombreUsuario,
		NumeroParticipacion: p.NumeroParticipacion,
		Origen:              p.Origen,
		FechaParticipacion:  p.FechaParticipacion,
	}
}

func ToParticipacionResponses(participaciones []ParticipacionRifa) []*ParticipacionResponse {
	responses := make([]*ParticipacionResponse, len(participaciones))
	for i, p := range participaciones {
		responses[i] = ToParticipacionResponse(&p)
	}
	return responses
}

func ToGanadorResponse(g *GanadorRifa) *GanadorResponse {
	numero := ""
	var usuarioID uint
	nombreGanador := ""
	if g.Participacion != nil {
		numero = g.Participacion.NumeroParticipacion
		usuarioID = g.Participacion.UsuarioID
		if g.Participacion.Usuario != nil {
			nombreGanador = g.Participacion.Usuario.FirstName + " " + g.Participacion.Usuario.LastName
		}
	}
	return &GanadorResponse{
		ID:                  g.ID,
		RifaID:              g.RifaID,
		ParticipacionID:     g.ParticipacionID,
		NumeroParticipacion: numero,
		UsuarioID:           usuarioID,
		NombreGanador:       nombreGanador,
		FechaNotificacion:   g.FechaNotificacion,
		EstadoEntrega:       g.EstadoEntrega,
	}
}

func ToGanadorResponses(ganadores []GanadorRifa) []*GanadorResponse {
	responses := make([]*GanadorResponse, len(ganadores))
	for i, g := range ganadores {
		responses[i] = ToGanadorResponse(&g)
	}
	return responses
}
