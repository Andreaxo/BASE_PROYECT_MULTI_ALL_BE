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

	esPremioMayor := false
	if req.EsPremioMayor != nil {
		esPremioMayor = *req.EsPremioMayor
	}

	rifa := &domain.Rifa{
		Nombre:        req.Nombre,
		Descripcion:   req.Descripcion,
		Premio:        req.Premio,
		ImagenURL:     req.ImagenURL,
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		FechaSorteo:   fechaSorteo,
		EsPremioMayor: esPremioMayor,
		Estado:        domain.EstadoActiva,
		CreateBy:      createdBy,
	}

	if err := s.rifaRepo.Create(rifa); err != nil {
		return nil, err
	}

	// Auto-enroll all active users with their base ticket
	_ = s.AsegurarInscripcionTodosLosUsuarios(rifa.ID)

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

	if req.EsPremioMayor != nil {
		rifa.EsPremioMayor = *req.EsPremioMayor
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

	// Ensure all registered users are enrolled in this raffle
	_ = s.AsegurarInscripcionTodosLosUsuarios(rifa.ID)

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

	// Ensure all registered users are enrolled with at least one base ticket
	_ = s.AsegurarInscripcionTodosLosUsuarios(id)

	participaciones, err := s.participacionRepo.FindByRifaID(id)
	if err != nil || len(participaciones) == 0 {
		return nil, errors.New("no hay participaciones para sortear esta rifa")
	}

	// Pick a winner if no winner registered yet
	ganadores, _ := s.ganadorRepo.FindByRifaID(id)
	if len(ganadores) == 0 {
		var pGanadora domain.ParticipacionRifa

		if !rifa.EsPremioMayor {
			// Sorteo estándar: probabilidad equitativa para cada participación
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := r.Intn(len(participaciones))
			pGanadora = participaciones[idx]
		} else {
			// Sorteo de PREMIO MAYOR:
			// Las personas con más referidos en el periodo de juego tienen más oportunidades de ganar.
			// 1. Rango de corte: desde el último premio mayor sorteado (o fecha_inicio) hasta la fecha actual
			var ultimoPremioMayor domain.Rifa
			err := s.db.Where("es_premio_mayor = ? AND estado = ? AND id != ?", true, domain.EstadoSorteada, id).
				Order("COALESCE(update_at, create_at) DESC, id DESC").
				First(&ultimoPremioMayor).Error

			now := time.Now()
			fechaCorteFin := now

			var ultimoPtr *domain.Rifa
			if err == nil && ultimoPremioMayor.ID != 0 {
				ultimoPtr = &ultimoPremioMayor
			}
			fechaCorteInicio := determineFechaCorteInicio(*rifa, ultimoPtr, now)

			// 2. Contar referidos generados por cada participante dentro de este periodo de juego
			type UserRefCount struct {
				UsuarioReferenteID uint
				Total              int64
			}
			var refCounts []UserRefCount
			s.db.Table("administrative.referido").
				Select("usuario_referente_id, COUNT(*) as total").
				Where("fecha_referido >= ? AND fecha_referido <= ?", fechaCorteInicio, fechaCorteFin).
				Group("usuario_referente_id").
				Scan(&refCounts)

			refMap := make(map[uint]int64)
			for _, rc := range refCounts {
				refMap[rc.UsuarioReferenteID] = rc.Total
			}

			// 3. Crear pool ponderado: cada participación base aporta 1 oportunidad + 1 oportunidad adicional por cada referido en el período
			var weightedPool []domain.ParticipacionRifa
			for _, p := range participaciones {
				// Boleto base
				weightedPool = append(weightedPool, p)
				// Oportunidades extras por referidos en el período de juego
				extraChances := refMap[p.UsuarioID]
				for i := int64(0); i < extraChances; i++ {
					weightedPool = append(weightedPool, p)
				}
			}

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := r.Intn(len(weightedPool))
			pGanadora = weightedPool[idx]
		}

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
	if !now.Before(rifa.FechaFin) {
		rifa.FechaSorteo = now
	}
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

		// Only user and user_member roles can participate in raffles
		var roleCode string
		_ = s.db.Table("administrative.users u").
			Joins("JOIN administrative.roles r ON u.role_id = r.id").
			Where("u.id = ?", usuarioID).
			Pluck("r.code", &roleCode).Error
		if roleCode != "user" && roleCode != "user_member" {
			return errors.New("solo los usuarios y afiliados pueden participar en las rifas")
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
	// Auto-enroll all active users if the rifa is active
	var rifa domain.Rifa
	if err := s.db.Table("administrative.rifa").Where("id = ?", rifaID).First(&rifa).Error; err == nil {
		if rifa.Estado == domain.EstadoActiva {
			_ = s.AsegurarInscripcionTodosLosUsuarios(rifaID)
		}
	}

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

	// Auto-enroll user if they are an eligible consumer/affiliate ('user' or 'user_member')
	var roleCode string
	_ = s.db.Table("administrative.users u").
		Joins("JOIN administrative.roles r ON u.role_id = r.id").
		Where("u.id = ?", usuarioID).
		Pluck("r.code", &roleCode).Error

	if roleCode == "user" || roleCode == "user_member" {
		var userTickets int64
		_ = s.db.Table("administrative.participacion_rifa").
			Where("rifa_id = ? AND usuario_id = ?", rifaActiva.ID, usuarioID).
			Count(&userTickets).Error
		if userTickets == 0 {
			var maxNum int
			row := s.db.Raw("SELECT COALESCE(MAX(NULLIF(regexp_replace(numero_participacion, '[^0-9]', '', 'g'), '')::integer), 0) FROM administrative.participacion_rifa WHERE rifa_id = ?", rifaActiva.ID).Row()
			_ = row.Scan(&maxNum)
			numTicket := fmt.Sprintf("%06d", maxNum+1)
			_ = s.db.Exec(`
				INSERT INTO administrative.participacion_rifa (rifa_id, usuario_id, numero_participacion, origen, fecha_participacion)
				VALUES (?, ?, ?, 'afiliacion', NOW())
			`, rifaActiva.ID, usuarioID, numTicket)
		}
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

func (s *rifaService) AsegurarInscripcionTodosLosUsuarios(rifaID uint) error {
	var userIDs []uint
	// Fetch only active users with consumer/affiliate roles (user, user_member), excluding administrative roles
	if err := s.db.Table("administrative.users u").
		Joins("JOIN administrative.roles r ON u.role_id = r.id").
		Where("u.is_active = true AND r.code IN ('user', 'user_member')").
		Pluck("u.id", &userIDs).Error; err != nil {
		return err
	}

	var enrolledUserIDs []uint
	_ = s.db.Table("administrative.participacion_rifa").
		Where("rifa_id = ?", rifaID).
		Pluck("usuario_id", &enrolledUserIDs).Error

	enrolledMap := make(map[uint]bool)
	for _, id := range enrolledUserIDs {
		enrolledMap[id] = true
	}

	var maxNum int
	row := s.db.Raw("SELECT COALESCE(MAX(NULLIF(regexp_replace(numero_participacion, '[^0-9]', '', 'g'), '')::integer), 0) FROM administrative.participacion_rifa WHERE rifa_id = ?", rifaID).Row()
	_ = row.Scan(&maxNum)

	for _, uid := range userIDs {
		if !enrolledMap[uid] {
			maxNum++
			numTicket := fmt.Sprintf("%06d", maxNum)
			_ = s.db.Exec(`
				INSERT INTO administrative.participacion_rifa (rifa_id, usuario_id, numero_participacion, origen, fecha_participacion)
				VALUES (?, ?, ?, 'afiliacion', NOW())
			`, rifaID, uid, numTicket)
		}
	}

	return nil
}

// determineFechaCorteInicio computes the starting cutoff timestamp for calculating referral tickets in a Premio Mayor.
// It prioritizes the previous Premio Mayor draw completion date, normalizes to the start of the day to prevent
// hour/minute discrepancies, and guards against scheduled future dates.
func determineFechaCorteInicio(rifa domain.Rifa, ultimoPremioMayor *domain.Rifa, now time.Time) time.Time {
	var fechaCorteInicio time.Time
	if ultimoPremioMayor != nil && ultimoPremioMayor.ID != 0 {
		var prevDrawDate time.Time
		if ultimoPremioMayor.UpdateAt != nil {
			prevDrawDate = *ultimoPremioMayor.UpdateAt
		} else {
			prevDrawDate = ultimoPremioMayor.CreateAt
		}

		if !prevDrawDate.After(now) {
			if rifa.FechaInicio.Before(prevDrawDate) {
				fechaCorteInicio = prevDrawDate
			} else {
				fechaCorteInicio = rifa.FechaInicio
			}
		} else {
			fechaCorteInicio = rifa.FechaInicio
		}
	} else {
		fechaCorteInicio = rifa.FechaInicio
	}

	// Normalizar al inicio del día (00:00:00) para no recortar referidos generados en el mismo día
	fechaCorteInicio = time.Date(fechaCorteInicio.Year(), fechaCorteInicio.Month(), fechaCorteInicio.Day(), 0, 0, 0, 0, fechaCorteInicio.Location())

	// Salvaguarda: Si por fechas programadas en el formulario fechaCorteInicio es posterior a ahora,
	// se ajusta al inicio del día actual para asegurar que la ventana sea válida (fechaCorteInicio <= now).
	if fechaCorteInicio.After(now) {
		fechaCorteInicio = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	return fechaCorteInicio
}

