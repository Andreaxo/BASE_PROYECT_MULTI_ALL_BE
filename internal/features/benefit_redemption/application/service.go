package application

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"multicliente-backend/internal/features/benefit_redemption/domain"
)

type redemptionService struct {
	repo domain.RedemptionRepository
}

func NewRedemptionService(repo domain.RedemptionRepository) domain.RedemptionService {
	return &redemptionService{repo: repo}
}

// RedeemBenefit generates a validation code for an authenticated employee.
func (s *redemptionService) RedeemBenefit(benefitID uint, userID uint) (*domain.RedemptionResponse, error) {
	// 1. Validate benefit exists and is active
	benefit, err := s.repo.FindBenefitByID(benefitID)
	if err != nil {
		return nil, errors.New("beneficio no encontrado")
	}

	if benefit.IsActive != nil && !*benefit.IsActive {
		return nil, errors.New("este beneficio no está activo")
	}

	estadoLower := strings.ToLower(strings.TrimSpace(benefit.Estado))
	if estadoLower == "inactivo" || estadoLower == "false" || estadoLower == "0" {
		return nil, errors.New("este beneficio no está activo")
	}

	// 2. Validate date range
	now := time.Now()
	if benefit.FechaInicio != nil && now.Before(*benefit.FechaInicio) {
		return nil, errors.New("este beneficio aún no está disponible")
	}
	if benefit.FechaFin != nil && now.After(benefit.FechaFin.AddDate(0, 0, 1)) {
		return nil, errors.New("este beneficio ya venció")
	}

	// 3. Generate code and attempt insert (with retry for code collision)
	const maxRetries = 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := generateValidationCode()
		if err != nil {
			return nil, errors.New("error generando código de validación")
		}

		redemption := &domain.BenefitRedemption{
			BenefitID:        benefitID,
			UsuarioID:        userID,
			CodigoValidacion: code,
			Estado:           domain.EstadoGenerado,
		}

		err = s.repo.Create(redemption)
		if err == nil {
			// Success — load benefit name for response
			redemption.Benefit = benefit
			return domain.ToRedemptionResponse(redemption), nil
		}

		errMsg := err.Error()

		// Check if it's a duplicate on (benefit_id, usuario_id) — business rule violation
		if strings.Contains(errMsg, "uq_redemption_benefit_usuario") {
			return nil, errors.New("ya has redimido este beneficio anteriormente")
		}

		// Check if it's a duplicate on codigo_validacion — retry with new code
		if strings.Contains(errMsg, "codigo_validacion") || strings.Contains(errMsg, "benefit_redemption_codigo_validacion_key") {
			continue
		}

		// Any other DB error — fail immediately
		return nil, errors.New("error al crear la redención")
	}

	return nil, errors.New("no se pudo generar un código único, intenta de nuevo")
}

// ValidateCode validates a redemption code from a business_validator user.
func (s *redemptionService) ValidateCode(code string, negocioUserID uint) (*domain.ValidateCodeResponse, error) {
	// 1. Find redemption by code
	redemption, err := s.repo.FindByCode(code)
	if err != nil {
		return nil, errors.New("código inválido")
	}

	// 2. Check expiration inline: if benefit has ended and code is still 'generado', mark as expired
	if redemption.Benefit != nil && redemption.Benefit.FechaFin != nil {
		now := time.Now()
		if now.After(redemption.Benefit.FechaFin.AddDate(0, 0, 1)) && redemption.Estado == domain.EstadoGenerado {
			// Atomic update — only marks if still 'generado'
			s.repo.MarkAsExpired(redemption.ID)
			return nil, errors.New("código vencido — el beneficio ya expiró")
		}
	}

	// 3. Validate current estado
	if redemption.Estado == domain.EstadoUsado {
		return nil, errors.New("código ya utilizado")
	}
	if redemption.Estado == domain.EstadoVencido {
		return nil, errors.New("código vencido")
	}
	if redemption.Estado != domain.EstadoGenerado {
		return nil, errors.New("código en estado inválido")
	}

	// 4. Validate company ownership — the negocio user must belong to the same company as the benefit
	empresaID, err := s.repo.GetUserEmpresaID(negocioUserID)
	if err != nil || empresaID == 0 {
		return nil, errors.New("no se pudo verificar tu empresa — contacta al administrador")
	}

	if redemption.Benefit == nil || redemption.Benefit.CompanyBenefit == nil {
		return nil, errors.New("el beneficio no tiene empresa asociada")
	}

	if *redemption.Benefit.CompanyBenefit != empresaID {
		return nil, errors.New("este código no pertenece a tu empresa")
	}

	// 5. Atomic update — only succeeds if estado is still 'generado'
	rowsAffected, err := s.repo.MarkAsUsed(redemption.ID, negocioUserID)
	if err != nil {
		return nil, errors.New("error al validar el código")
	}
	if rowsAffected == 0 {
		return nil, errors.New("el código ya fue procesado por otra operación")
	}

	// 6. Build response — get user name from the redemption
	now := time.Now()
	return &domain.ValidateCodeResponse{
		ID:          redemption.ID,
		BenefitName: redemption.Benefit.Name,
		Estado:      domain.EstadoUsado,
		FechaUso:    &now,
	}, nil
}

func (s *redemptionService) GetMisRedenciones(userID uint) ([]*domain.RedemptionResponse, error) {
	redemptions, err := s.repo.FindByUsuarioID(userID)
	if err != nil {
		return nil, err
	}
	return domain.ToRedemptionResponses(redemptions), nil
}

func (s *redemptionService) GetRedemptionsByEmpresa(negocioUserID uint) ([]*domain.RedemptionResponse, error) {
	empresaID, err := s.repo.GetUserEmpresaID(negocioUserID)
	if err != nil || empresaID == 0 {
		return nil, errors.New("no se pudo determinar tu empresa")
	}

	redemptions, err := s.repo.FindByEmpresaID(empresaID)
	if err != nil {
		return nil, err
	}
	return domain.ToRedemptionResponses(redemptions), nil
}

func (s *redemptionService) GetAllRedemptions() ([]*domain.RedemptionResponse, error) {
	redemptions, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return domain.ToRedemptionResponses(redemptions), nil
}

// generateValidationCode generates a random 8-character code using a reduced charset
// without ambiguous characters (0/O, 1/I/L removed).
// Uses crypto/rand for security — codes must not be guessable or sequential.
func generateValidationCode() (string, error) {
	code := make([]byte, domain.CodeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(domain.CodeCharset))))
		if err != nil {
			return "", err
		}
		code[i] = domain.CodeCharset[n.Int64()]
	}
	return string(code), nil
}
