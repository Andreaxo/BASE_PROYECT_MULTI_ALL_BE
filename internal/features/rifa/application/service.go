package application

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

	referidoDomain "multicliente-backend/internal/features/referido/domain"
	"multicliente-backend/internal/features/rifa/domain"
)

type rifaService struct {
	db                *gorm.DB
	rifaRepo          domain.RifaRepository
	participacionRepo domain.ParticipacionRepository
	ganadorRepo       domain.GanadorRepository
	referidoRepo      referidoDomain.ReferidoRepository
}

func NewRifaService(
	db *gorm.DB,
	rifaRepo domain.RifaRepository,
	participacionRepo domain.ParticipacionRepository,
	ganadorRepo domain.GanadorRepository,
	referidoRepo referidoDomain.ReferidoRepository,
) domain.RifaService {
	return &rifaService{
		db:                db,
		rifaRepo:          rifaRepo,
		participacionRepo: participacionRepo,
		ganadorRepo:       ganadorRepo,
		referidoRepo:      referidoRepo,
	}
}

func (s *rifaService) CreateRifa(req *domain.CreateRifaRequest, createdBy *uint) (*domain.RifaResponse, error) {
	fechaInicio, err := time.Parse("2006-01-02", req.FechaInicio)
	if err != nil {
		return nil, errors.New("fecha_inicio invalida, debe ser YYYY-MM-DD")
	}
	fechaFin, err := time.Parse("2006-01-02", req.FechaFin)
	if err != nil {
		return nil, errors.New("fecha_fin invalida, debe ser YYYY-MM-DD")
	}
	fechaSorteo, err := time.Parse("2006-01-02", req.FechaSorteo)
	if err != nil {
		return nil, errors.New("fecha_sorteo invalida, debe ser YYYY-MM-DD")
	}

	rifa := &domain.Rifa{
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
		Premio:      req.Premio,
		ImagenURL:   req.ImagenURL,
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
		FechaSorteo: fechaSorteo,
		Estado:      domain.EstadoActiva,
		CreateBy:    createdBy,
	}

	if err := s.rifaRepo.Create(rifa); err != nil {
		return nil, err
	}

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) UpdateRifa(id uint, req *domain.UpdateRifaRequest, updatedBy *uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}

	if req.Nombre != nil {
		rifa.Nombre = *req.Nombre
	}
	if req.Descripcion != nil {
		rifa.Descripcion = req.Descripcion
	}
	if req.Premio != nil {
		rifa.Premio = *req.Premio
	}
	if req.ImagenURL != nil {
		rifa.ImagenURL = req.ImagenURL
	}
	if req.FechaInicio != nil {
		fi, err := time.Parse("2006-01-02", *req.FechaInicio)
		if err == nil {
			rifa.FechaInicio = fi
		}
	}
	if req.FechaFin != nil {
		ff, err := time.Parse("2006-01-02", *req.FechaFin)
		if err == nil {
			rifa.FechaFin = ff
		}
	}
	if req.FechaSorteo != nil {
		fs, err := time.Parse("2006-01-02", *req.FechaSorteo)
		if err == nil {
			rifa.FechaSorteo = fs
		}
	}

	now := time.Now()
	rifa.UpdateBy = updatedBy
	rifa.UpdateAt = &now

	if err := s.rifaRepo.Update(rifa); err != nil {
		return nil, err
	}

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) ActivarRifa(id uint, updatedBy *uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}

	now := time.Now()
	rifa.Estado = domain.EstadoActiva
	rifa.UpdateBy = updatedBy
	rifa.UpdateAt = &now

	if err := s.rifaRepo.Update(rifa); err != nil {
		return nil, err
	}

	// Grant pending rewards
	_ = s.OtorgarRecompensasPendientes(rifa.ID)

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) CerrarRifa(id uint, updatedBy *uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}

	now := time.Now()
	rifa.Estado = domain.EstadoCerrada
	rifa.UpdateBy = updatedBy
	rifa.UpdateAt = &now

	if err := s.rifaRepo.Update(rifa); err != nil {
		return nil, err
	}

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) SortearRifa(id uint, updatedBy *uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}

	participaciones, err := s.participacionRepo.FindByRifaID(id)
	if err != nil || len(participaciones) == 0 {
		return nil, errors.New("no hay participaciones para sortear esta rifa")
	}

	// Pick a random winner if no winner registered yet
	ganadores, _ := s.ganadorRepo.FindByRifaID(id)
	if len(ganadores) == 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		idx := r.Intn(len(participaciones))
		pGanadora := participaciones[idx]

		now := time.Now()
		g := &domain.GanadorRifa{
			RifaID:            id,
			ParticipacionID:   pGanadora.ID,
			FechaNotificacion: &now,
			EstadoEntrega:     domain.EntregaPendiente,
		}
		_ = s.ganadorRepo.Create(g)
	}

	now := time.Now()
	rifa.Estado = domain.EstadoSorteada
	rifa.UpdateBy = updatedBy
	rifa.UpdateAt = &now

	if err := s.rifaRepo.Update(rifa); err != nil {
		return nil, err
	}

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) CancelarRifa(id uint, updatedBy *uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}

	now := time.Now()
	rifa.Estado = domain.EstadoCancelada
	rifa.UpdateBy = updatedBy
	rifa.UpdateAt = &now

	if err := s.rifaRepo.Update(rifa); err != nil {
		return nil, err
	}

	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) GetRifaActiva() (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindActiva()
	if err != nil {
		return nil, errors.New("no hay rifa activa")
	}
	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) GetRifaByID(id uint) (*domain.RifaResponse, error) {
	rifa, err := s.rifaRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("rifa no encontrada")
	}
	return domain.ToRifaResponse(rifa), nil
}

func (s *rifaService) GetAllRifas() ([]*domain.RifaResponse, error) {
	rifas, err := s.rifaRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return domain.ToRifaResponses(rifas), nil
}

func (s *rifaService) AgregarParticipacion(rifaID uint, usuarioID uint, origen string) (*domain.ParticipacionResponse, error) {
	var createdParticipacion *domain.ParticipacionRifa

	err := s.db.Transaction(func(tx *gorm.DB) error {
		rifa, err := s.rifaRepo.FindByIDForUpdate(tx, rifaID)
		if err != nil {
			return errors.New("rifa no encontrada")
		}

		if rifa.Estado != domain.EstadoActiva {
			return errors.New("la rifa no esta activa")
		}

		if origen == domain.OrigenManual {
			existing, _ := s.participacionRepo.FindByUsuarioAndRifa(usuarioID, rifaID)
			for _, item := range existing {
				if item.Origen == domain.OrigenManual {
					return errors.New("Ya estás participando en esta rifa")
				}
			}
		}

		count, err := s.participacionRepo.CountByRifaTx(tx, rifaID)
		if err != nil {
			return err
		}

		numTicket := fmt.Sprintf("%06d", count+1)
		p := &domain.ParticipacionRifa{
			RifaID:              rifaID,
			UsuarioID:           usuarioID,
			NumeroParticipacion: numTicket,
			Origen:              origen,
		}

		if err := s.participacionRepo.CreateTx(tx, p); err != nil {
			return err
		}

		createdParticipacion = p
		return nil
	})

	if err != nil {
		return nil, err
	}

	full, err := s.participacionRepo.FindByID(createdParticipacion.ID)
	if err != nil {
		return domain.ToParticipacionResponse(createdParticipacion), nil
	}

	return domain.ToParticipacionResponse(full), nil
}

func (s *rifaService) GetParticipacionesByRifa(rifaID uint) ([]*domain.ParticipacionResponse, error) {
	participaciones, err := s.participacionRepo.FindByRifaID(rifaID)
	if err != nil {
		return nil, err
	}
	return domain.ToParticipacionResponses(participaciones), nil
}

func (s *rifaService) GetMisParticipaciones(usuarioID uint) ([]*domain.ParticipacionResponse, error) {
	rifaActiva, err := s.rifaRepo.FindActiva()
	if err != nil {
		return []*domain.ParticipacionResponse{}, nil
	}

	participaciones, err := s.participacionRepo.FindByUsuarioAndRifa(usuarioID, rifaActiva.ID)
	if err != nil {
		return nil, err
	}
	return domain.ToParticipacionResponses(participaciones), nil
}

func (s *rifaService) RegistrarGanador(rifaID uint, participacionID uint) (*domain.GanadorResponse, error) {
	p, err := s.participacionRepo.FindByID(participacionID)
	if err != nil {
		return nil, errors.New("participacion no encontrada")
	}

	if p.RifaID != rifaID {
		return nil, errors.New("la participacion no pertenece a esta rifa")
	}

	now := time.Now()
	g := &domain.GanadorRifa{
		RifaID:            rifaID,
		ParticipacionID:   participacionID,
		FechaNotificacion: &now,
		EstadoEntrega:     domain.EntregaPendiente,
	}

	if err := s.ganadorRepo.Create(g); err != nil {
		return nil, err
	}

	full, err := s.ganadorRepo.FindByID(g.ID)
	if err != nil {
		return domain.ToGanadorResponse(g), nil
	}

	return domain.ToGanadorResponse(full), nil
}

func (s *rifaService) ActualizarEstadoEntrega(ganadorID uint, nuevoEstado string) (*domain.GanadorResponse, error) {
	g, err := s.ganadorRepo.FindByID(ganadorID)
	if err != nil {
		return nil, errors.New("ganador no encontrado")
	}

	g.EstadoEntrega = nuevoEstado
	if err := s.ganadorRepo.Update(g); err != nil {
		return nil, err
	}

	full, err := s.ganadorRepo.FindByID(g.ID)
	if err != nil {
		return domain.ToGanadorResponse(g), nil
	}

	return domain.ToGanadorResponse(full), nil
}

func (s *rifaService) GetGanadoresByRifa(rifaID uint) ([]*domain.GanadorResponse, error) {
	ganadores, err := s.ganadorRepo.FindByRifaID(rifaID)
	if err != nil {
		return nil, err
	}
	return domain.ToGanadorResponses(ganadores), nil
}

func (s *rifaService) GetHistorial() ([]*domain.RifaHistorialResponse, error) {
	rifas, err := s.rifaRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var historial []*domain.RifaHistorialResponse
	for _, r := range rifas {
		if r.Estado == domain.EstadoCerrada || r.Estado == domain.EstadoSorteada {
			ganadores, _ := s.ganadorRepo.FindByRifaID(r.ID)
			gResponses := domain.ToGanadorResponses(ganadores)
			var gValues []domain.GanadorResponse
			for _, g := range gResponses {
				gValues = append(gValues, *g)
			}
			historial = append(historial, &domain.RifaHistorialResponse{
				Rifa:      *domain.ToRifaResponse(&r),
				Ganadores: gValues,
			})
		}
	}

	return historial, nil
}

func (s *rifaService) OtorgarRecompensaReferido(referenteID uint) error {
	rifaActiva, err := s.rifaRepo.FindActiva()
	if err != nil {
		return errors.New("no hay rifa activa para otorgar recompensa")
	}

	_, err = s.AgregarParticipacion(rifaActiva.ID, referenteID, domain.OrigenReferido)
	return err
}

func (s *rifaService) OtorgarRecompensasPendientes(rifaID uint) error {
	if s.referidoRepo == nil {
		return nil
	}

	pendientes, err := s.referidoRepo.FindPendingRewards()
	if err != nil {
		return err
	}

	for _, ref := range pendientes {
		_, err := s.AgregarParticipacion(rifaID, ref.UsuarioReferenteID, domain.OrigenReferido)
		if err == nil {
			ref.RecompensaOtorgada = true
			_ = s.referidoRepo.Update(&ref)
		}
	}

	return nil
}
