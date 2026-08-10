package application

import (
	"errors"
	"time"

	"multicliente-backend/internal/features/referido/domain"
)

type referidoService struct {
	repo domain.ReferidoRepository
}

// NewReferidoService creates a new ReferidoService with the given repository.
func NewReferidoService(repo domain.ReferidoRepository) domain.ReferidoService {
	return &referidoService{repo: repo}
}

// CreateReferido registers a new referral invitation in 'pendiente' state.
func (s *referidoService) CreateReferido(req *domain.CreateReferidoRequest, referenteID uint) (*domain.ReferidoResponse, error) {
	// Check if this email was already referred
	existing, _ := s.repo.FindByEmail(req.EmailReferido)
	if existing != nil {
		return nil, errors.New("este correo ya fue referido anteriormente")
	}

	referido := &domain.Referido{
		UsuarioReferenteID: referenteID,
		EmailReferido:      req.EmailReferido,
		NombreReferido:     req.NombreReferido,
		Estado:             domain.EstadoPendiente,
	}

	if err := s.repo.Create(referido); err != nil {
		return nil, err
	}

	// Reload with relations
	full, err := s.repo.FindByID(referido.ID)
	if err != nil {
		return nil, err
	}

	return domain.ToReferidoResponse(full), nil
}

func (s *referidoService) GetReferidoByID(id uint) (*domain.ReferidoResponse, error) {
	referido, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("referido no encontrado")
	}
	return domain.ToReferidoResponse(referido), nil
}

func (s *referidoService) GetReferidosByReferente(referenteID uint) ([]*domain.ReferidoResponse, error) {
	referidos, err := s.repo.FindByReferenteID(referenteID)
	if err != nil {
		return nil, err
	}
	return domain.ToReferidoResponses(referidos), nil
}

func (s *referidoService) GetAllReferidos() ([]*domain.ReferidoResponse, error) {
	referidos, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return domain.ToReferidoResponses(referidos), nil
}

// ConvertirReferido transitions a referral from 'pendiente' to 'registrado'.
// Called when the referred user creates their account.
func (s *referidoService) ConvertirReferido(email string, usuarioReferidoID uint) (*domain.ReferidoResponse, error) {
	referido, err := s.repo.FindPendingByEmail(email)
	if err != nil {
		return nil, errors.New("no se encontró una invitación pendiente para este correo")
	}

	// Prevent self-referral
	if referido.UsuarioReferenteID == usuarioReferidoID {
		return nil, errors.New("un usuario no puede referirse a sí mismo")
	}

	referido.UsuarioReferidoID = &usuarioReferidoID
	referido.Estado = domain.EstadoRegistrado

	if err := s.repo.Update(referido); err != nil {
		return nil, err
	}

	full, err := s.repo.FindByID(referido.ID)
	if err != nil {
		return nil, err
	}
	return domain.ToReferidoResponse(full), nil
}

// AfiliarReferido transitions a referral from 'registrado' to 'afiliado'.
// Sets the fecha_conversion timestamp.
func (s *referidoService) AfiliarReferido(id uint) (*domain.ReferidoResponse, error) {
	referido, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("referido no encontrado")
	}

	if referido.Estado != domain.EstadoRegistrado {
		return nil, errors.New("solo se pueden afiliar referidos en estado 'registrado'")
	}

	now := time.Now()
	referido.Estado = domain.EstadoAfiliado
	referido.FechaConversion = &now

	if err := s.repo.Update(referido); err != nil {
		return nil, err
	}

	full, err := s.repo.FindByID(referido.ID)
	if err != nil {
		return nil, err
	}
	return domain.ToReferidoResponse(full), nil
}

// OtorgarRecompensa marks the reward as granted for an affiliated referral.
// The actual reward logic (e.g., raffle entries) should be integrated here.
func (s *referidoService) OtorgarRecompensa(id uint) (*domain.ReferidoResponse, error) {
	referido, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("referido no encontrado")
	}

	if referido.Estado != domain.EstadoAfiliado {
		return nil, errors.New("solo se puede otorgar recompensa a referidos en estado 'afiliado'")
	}

	if referido.RecompensaOtorgada {
		return nil, errors.New("la recompensa ya fue otorgada para este referido")
	}

	// TODO: Here goes the actual reward logic (e.g., grant raffle participation).
	// This is a placeholder for future integration.

	referido.RecompensaOtorgada = true

	if err := s.repo.Update(referido); err != nil {
		return nil, err
	}

	full, err := s.repo.FindByID(referido.ID)
	if err != nil {
		return nil, err
	}
	return domain.ToReferidoResponse(full), nil
}

func (s *referidoService) DeleteReferido(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("referido no encontrado")
	}
	return s.repo.Delete(id)
}
