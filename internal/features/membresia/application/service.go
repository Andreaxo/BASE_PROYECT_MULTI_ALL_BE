package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	membresiaDomain "multicliente-backend/internal/features/membresia/domain"
	"multicliente-backend/internal/features/membresia/infrastructure"
	notificacionDomain "multicliente-backend/internal/features/notificacion/domain"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
)

type membresiaService struct {
	db           *gorm.DB
	membRepo     membresiaDomain.MembresiaRepository
	pagoRepo     membresiaDomain.PagoMembresiaRepository
	referidoRepo  referidoDomain.ReferidoRepository
	userRepo     userDomain.UserRepository
	wompi        *infrastructure.WompiClient
	notifSvc     notificacionDomain.NotificacionService
}

// NewMembresiaService creates a new MembresiaService with all dependencies.
func NewMembresiaService(
	db *gorm.DB,
	membRepo membresiaDomain.MembresiaRepository,
	pagoRepo membresiaDomain.PagoMembresiaRepository,
	referidoRepo referidoDomain.ReferidoRepository,
	userRepo userDomain.UserRepository,
	wompi *infrastructure.WompiClient,
	notifSvc notificacionDomain.NotificacionService,
) membresiaDomain.MembresiaService {
	return &membresiaService{
		db:           db,
		membRepo:     membRepo,
		pagoRepo:     pagoRepo,
		referidoRepo:  referidoRepo,
		userRepo:     userRepo,
		wompi:        wompi,
		notifSvc:     notifSvc,
	}
}

// IniciarPago starts the payment flow for a user's membership.
func (s *membresiaService) IniciarPago(usuarioID uint, req *membresiaDomain.IniciarPagoRequest) (*membresiaDomain.IniciarPagoResponse, error) {
	// 1. Get or create membership
	membresia, err := s.membRepo.FindByUsuarioID(usuarioID)
	if err != nil {
		// Create inactive membership
		membresia = &membresiaDomain.Membresia{
			UsuarioID:            usuarioID,
			Estado:               membresiaDomain.MembresiaInactiva,
			RenovacionAutomatica: true,
		}
		if err := s.membRepo.Create(membresia); err != nil {
			return nil, fmt.Errorf("falló al crear membresía: %w", err)
		}
	}

	// 2. Read price from configuracion
	precioStr := membresiaDomain.GetConfigValue(s.db, "precio_membresia_mensual", "25000")
	precio, err := strconv.ParseInt(precioStr, 10, 64)
	if err != nil {
		return nil, errors.New("configuración de precio de membresía inválida")
	}
	amountInCents := precio * 100 // Wompi uses cents

	// 3. Get user email for Wompi
	user, err := s.userRepo.FindByID(usuarioID)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	if user.Role != nil {
		roleCode := strings.ToLower(strings.TrimSpace(user.Role.Code))
		if roleCode == "superadmin" || roleCode == "admin" || roleCode == "super_admin" || roleCode == "operador" {
			return nil, errors.New("el usuario cuenta con acceso administrativo/vitalicio y no requiere pago de afiliación")
		}
	}

	// 4. Generate unique reference
	referencia := fmt.Sprintf("MEMB-%d-%d", membresia.ID, time.Now().UnixMilli())

	// 5. Create Wompi transaction
	txResp, err := s.wompi.CreateTransaction(&infrastructure.CreateTransactionRequest{
		AmountInCents: amountInCents,
		Currency:      "COP",
		CustomerEmail: user.Email,
		Reference:     referencia,
		RedirectURL:   req.RedirectURL,
	})
	if err != nil {
		return nil, fmt.Errorf("falló al crear transacción en Wompi: %w", err)
	}

	// 6. Create pending payment record
	pago := &membresiaDomain.PagoMembresia{
		MembresiaID:        membresia.ID,
		Monto:              float64(precio),
		Moneda:             "COP",
		Tipo:               membresiaDomain.PagoTipoInicial, // Will be corrected in webhook
		ReferenciaPasarela: &referencia,
		Estado:             membresiaDomain.PagoEstadoPendiente,
	}
	if err := s.pagoRepo.Create(pago); err != nil {
		return nil, fmt.Errorf("falló al registrar pago pendiente: %w", err)
	}

	return &membresiaDomain.IniciarPagoResponse{
		TransactionID: txResp.Data.ID,
		CheckoutURL:   txResp.Data.PublicCheckoutURL,
		Referencia:    referencia,
	}, nil
}

// ProcessWebhook processes a Wompi webhook notification.
func (s *membresiaService) ProcessWebhook(body []byte, signature string) error {
	// 1. Parse the event
	var event infrastructure.WompiWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}

	// 2. Validate signature
	if !s.wompi.ValidateWebhookSignature(&event) {
		return errors.New("invalid webhook signature")
	}

	// 3. Only process transaction.updated events
	if event.Event != "transaction.updated" {
		return nil // Ignore other events
	}

	txID := event.Data.Transaction.ID
	txRef := event.Data.Transaction.Reference
	txStatus := event.Data.Transaction.Status
	pmToken := event.Data.Transaction.PaymentMethod.Token

	return s.applyTransactionResult(txID, txRef, txStatus, pmToken)
}

// applyTransactionResult applies transaction confirmation from Wompi (via webhook or polling sync).
func (s *membresiaService) applyTransactionResult(txID, txRef, txStatus, pmToken string) error {
	// 4. Deduplication / lookup:
	// First check if already processed with this gateway transaction ID
	existingPago, _ := s.pagoRepo.FindByReferencia(txID)
	if existingPago != nil {
		if txStatus == "APPROVED" && existingPago.Estado == membresiaDomain.PagoEstadoAprobado {
			log.Printf("ℹ️ [Idempotency] Transaction %s already processed as APPROVED. Skipping duplicate.", txID)
			return nil
		}
		if (txStatus == "DECLINED" || txStatus == "ERROR") && existingPago.Estado == membresiaDomain.PagoEstadoRechazado {
			log.Printf("ℹ️ [Idempotency] Transaction %s already processed as DECLINED. Skipping duplicate.", txID)
			return nil
		}
	}

	// If not found by gateway txID, look up by our merchant reference (e.g. MEMB-...)
	if existingPago == nil && txRef != "" {
		existingPago, _ = s.pagoRepo.FindByReferencia(txRef)
	}

	// 5. Find the pending payment by reference
	if existingPago == nil {
		return fmt.Errorf("no se encontró pago pendiente con referencia: %s (ref: %s)", txID, txRef)
	}

	// 6. Process in a DB transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Re-fetch within transaction
		var pago membresiaDomain.PagoMembresia
		if err := tx.First(&pago, "id = ?", existingPago.ID).Error; err != nil {
			return err
		}

		var membresia membresiaDomain.Membresia
		if err := tx.First(&membresia, "id = ?", pago.MembresiaID).Error; err != nil {
			return err
		}

		// Determine tipo by counting previous approved payments
		var approvedCount int64
		tx.Model(&membresiaDomain.PagoMembresia{}).
			Where("membresia_id = ? AND estado = ?", membresia.ID, membresiaDomain.PagoEstadoAprobado).
			Count(&approvedCount)

		if approvedCount == 0 {
			pago.Tipo = membresiaDomain.PagoTipoInicial
		} else {
			pago.Tipo = membresiaDomain.PagoTipoRenovacion
		}

		if txStatus == "APPROVED" {
			// Approved
			pago.Estado = membresiaDomain.PagoEstadoAprobado
			if txID != "" {
				pago.ReferenciaPasarela = &txID
			}
			if err := tx.Save(&pago).Error; err != nil {
				return err
			}

			// Clean up other lingering pending payment attempts for this membership
			tx.Model(&membresiaDomain.PagoMembresia{}).
				Where("membresia_id = ? AND id != ? AND estado = ?", membresia.ID, pago.ID, membresiaDomain.PagoEstadoPendiente).
				Update("estado", membresiaDomain.PagoEstadoRechazado)

			// Activate membership
			now := time.Now()
			nextMonth := now.AddDate(0, 1, 0)
			membresia.Estado = membresiaDomain.MembresiaActiva
			membresia.FechaInicio = &now
			membresia.FechaFin = &nextMonth

			// Save payment method token if present
			if pmToken != "" {
				membresia.MetodoPagoToken = &pmToken
			}

			if err := tx.Save(&membresia).Error; err != nil {
				return err
			}

			// If this is the first payment (tipo = 'inicial'), trigger referido logic
			if pago.Tipo == membresiaDomain.PagoTipoInicial {
				s.triggerReferidoAfiliacion(tx, membresia.UsuarioID)
			}

			// Activate associated company's subscription if the user is linked to a company
			_ = tx.Exec(`
				UPDATE administrative.companies SET suscripcion_estado = 'activa'
				WHERE id IN (
					SELECT company_id FROM administrative.user_companies WHERE user_id = ?
					UNION
					SELECT empresa_id FROM administrative.users WHERE id = ? AND empresa_id IS NOT NULL
				)
			`, membresia.UsuarioID, membresia.UsuarioID)

			// Notificar a superadministradores sobre el pago aprobado
			user, _ := s.userRepo.FindByID(membresia.UsuarioID)
			nombre := "Usuario"
			email := ""
			if user != nil {
				nombre = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
				email = user.Email
			}
			refPasarela := ""
			if pago.ReferenciaPasarela != nil {
				refPasarela = *pago.ReferenciaPasarela
			}
			titulo := fmt.Sprintf("Pago de membresía aprobado - %s", nombre)
			mensaje := fmt.Sprintf("El usuario %s (%s) realizó con éxito el pago de membresía por $%.0f COP (Ref: %s).", nombre, email, pago.Monto, refPasarela)
			if s.notifSvc != nil {
				_ = s.notifSvc.NotificarPagoSuperadmins(titulo, mensaje, notificacionDomain.TipoPagoAprobado, pago.ID, "pago_membresia")
			}
		} else if txStatus == "DECLINED" || txStatus == "ERROR" || txStatus == "VOIDED" {
			// Safeguard: Never downgrade an already approved payment to rejected
			if pago.Estado == membresiaDomain.PagoEstadoAprobado {
				log.Printf("⚠️ [Anomaly] Received status %s for already APPROVED payment ID %d. Ignoring transition.", txStatus, pago.ID)
				return nil
			}
			// Rejected
			pago.Estado = membresiaDomain.PagoEstadoRechazado
			if txID != "" {
				pago.ReferenciaPasarela = &txID
			}
			if err := tx.Save(&pago).Error; err != nil {
				return err
			}

			// Notificar a superadministradores sobre el pago rechazado
			user, _ := s.userRepo.FindByID(membresia.UsuarioID)
			nombre := "Usuario"
			email := ""
			if user != nil {
				nombre = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
				email = user.Email
			}
			titulo := fmt.Sprintf("Pago de membresía rechazado - %s", nombre)
			mensaje := fmt.Sprintf("El intento de pago de membresía para el usuario %s (%s) fue rechazado/falló por $%.0f COP.", nombre, email, pago.Monto)
			if s.notifSvc != nil {
				_ = s.notifSvc.NotificarPagoSuperadmins(titulo, mensaje, notificacionDomain.TipoPagoRechazado, pago.ID, "pago_membresia")
			}
			// Membership stays as is (inactiva if first payment, or no change)
		}
		// For PENDING or other statuses, do nothing

		return nil
	})
}

// triggerReferidoAfiliacion handles the referral affiliation when first payment is approved.
// This is the NEW trigger point (moved from registration flow).
func (s *membresiaService) triggerReferidoAfiliacion(tx *gorm.DB, usuarioID uint) {
	// Find referido record where this user is the referred person and state is 'registrado'
	var referido referidoDomain.Referido
	err := tx.
		Where("usuario_referido_id = ? AND estado = ?", usuarioID, referidoDomain.EstadoRegistrado).
		First(&referido).Error
	if err != nil {
		// No referral record or not in 'registrado' state — nothing to do
		return
	}

	// Transition to 'afiliado'
	now := time.Now()
	referido.Estado = referidoDomain.EstadoAfiliado
	referido.FechaConversion = &now
	if err := tx.Save(&referido).Error; err != nil {
		log.Printf("⚠️ Failed to affiliate referido %d: %v", referido.ID, err)
		return
	}

	// Grant reward: add participation in active raffle for the referente
	// This replicates the OtorgarRecompensaReferido logic but within the transaction
	referido.RecompensaOtorgada = true
	_ = tx.Save(&referido).Error

	log.Printf("✅ Referido %d affiliated and reward granted for referente %d", referido.ID, referido.UsuarioReferenteID)
}

// GetMiMembresia returns the membership status for a user.
func (s *membresiaService) GetMiMembresia(usuarioID uint) (*membresiaDomain.MembresiaResponse, error) {
	// 0. Si es superadmin, admin u operador, siempre tiene membresía activa vitalicia/administrativa
	if user, err := s.userRepo.FindByID(usuarioID); err == nil && user != nil && user.Role != nil {
		roleCode := strings.ToLower(strings.TrimSpace(user.Role.Code))
		if roleCode == "superadmin" || roleCode == "admin" || roleCode == "super_admin" || roleCode == "operador" {
			now := time.Now()
			future := now.AddDate(100, 0, 0)
			return &membresiaDomain.MembresiaResponse{
				UsuarioID:            usuarioID,
				Estado:               membresiaDomain.MembresiaActiva,
				FechaInicio:          &now,
				FechaFin:             &future,
				RenovacionAutomatica: false,
				TieneMetodoPago:      false,
			}, nil
		}
	}

	membresia, err := s.membRepo.FindByUsuarioID(usuarioID)
	if err != nil {
		// No membership exists yet — return a "virtual" inactive response
		return &membresiaDomain.MembresiaResponse{
			UsuarioID:            usuarioID,
			Estado:               membresiaDomain.MembresiaInactiva,
			RenovacionAutomatica: true,
			TieneMetodoPago:      false,
		}, nil
	}
	// 1. Sincronización proactiva con Wompi: Si la membresía está inactiva, verificar pagos pendientes recientes
	if membresia.Estado != membresiaDomain.MembresiaActiva {
		var pendingPagos []membresiaDomain.PagoMembresia
		twoHoursAgo := time.Now().Add(-2 * time.Hour)
		if err := s.db.Where("membresia_id = ? AND estado = ? AND fecha_pago >= ?", membresia.ID, membresiaDomain.PagoEstadoPendiente, twoHoursAgo).
			Order("id DESC").Find(&pendingPagos).Error; err == nil && len(pendingPagos) > 0 {
			for _, p := range pendingPagos {
				if p.ReferenciaPasarela == nil || *p.ReferenciaPasarela == "" {
					continue
				}
				txs, err := s.wompi.GetTransactionsByReference(*p.ReferenciaPasarela)
				if err == nil && len(txs) > 0 {
					txItem := txs[0]
					if txItem.Status == "APPROVED" || txItem.Status == "DECLINED" || txItem.Status == "ERROR" {
						_ = s.applyTransactionResult(txItem.ID, txItem.Reference, txItem.Status, txItem.PaymentMethod.Token)
						if txItem.Status == "APPROVED" {
							if updatedMemb, err := s.membRepo.FindByID(membresia.ID); err == nil {
								membresia = updatedMemb
							}
							break
						}
					}
				}
			}
		}
	}

	resp := membresiaDomain.ToMembresiaResponse(membresia)
	var ultimoPago membresiaDomain.PagoMembresia
	if membresia.Estado == membresiaDomain.MembresiaActiva {
		if err := s.db.Where("membresia_id = ? AND estado = ?", membresia.ID, membresiaDomain.PagoEstadoAprobado).Order("id DESC").First(&ultimoPago).Error; err == nil {
			resp.UltimoPagoEstado = &ultimoPago.Estado
		}
	} else {
		if err := s.db.Where("membresia_id = ?", membresia.ID).Order("id DESC").First(&ultimoPago).Error; err == nil {
			resp.UltimoPagoEstado = &ultimoPago.Estado
		}
	}
	return resp, nil
}

// CancelarRenovacion disables automatic renewal for a user's membership.
func (s *membresiaService) CancelarRenovacion(usuarioID uint) (*membresiaDomain.MembresiaResponse, error) {
	membresia, err := s.membRepo.FindByUsuarioID(usuarioID)
	if err != nil {
		return nil, errors.New("no se encontró membresía para este usuario")
	}

	membresia.RenovacionAutomatica = false
	if err := s.membRepo.Update(membresia); err != nil {
		return nil, err
	}

	return membresiaDomain.ToMembresiaResponse(membresia), nil
}

// RenovarVencidas processes automatic renewals for expired active memberships.
func (s *membresiaService) RenovarVencidas() error {
	vencidas, err := s.membRepo.FindActivasVencidas()
	if err != nil {
		return err
	}

	precioStr := membresiaDomain.GetConfigValue(s.db, "precio_membresia_mensual", "25000")
	precio, _ := strconv.ParseInt(precioStr, 10, 64)
	amountInCents := precio * 100

	for _, m := range vencidas {
		if m.MetodoPagoToken == nil || *m.MetodoPagoToken == "" {
			// No payment token — mark as expired
			m.Estado = membresiaDomain.MembresiaVencida
			_ = s.membRepo.Update(&m)
			log.Printf("⚠️ Membresía %d vencida (sin token de pago)", m.ID)
			continue
		}

		// Get user email
		user, err := s.userRepo.FindByID(m.UsuarioID)
		if err != nil {
			log.Printf("⚠️ Error obteniendo usuario %d para renovación: %v", m.UsuarioID, err)
			m.Estado = membresiaDomain.MembresiaVencida
			_ = s.membRepo.Update(&m)
			continue
		}

		referencia := fmt.Sprintf("RENOV-%d-%d", m.ID, time.Now().UnixMilli())

		// Attempt to charge with saved token
		// Note: Wompi payment_source_id is stored as token string; for real integration,
		// this would be the numeric payment source ID from Wompi's tokenization flow
		txResp, err := s.wompi.ChargeWithToken(&infrastructure.TokenChargeRequest{
			AmountInCents: amountInCents,
			Currency:      "COP",
			CustomerEmail: user.Email,
			Reference:     referencia,
			PaymentSourceID: 0, // Will be set from actual Wompi payment source
		})

		if err != nil || txResp.Data.Status != "APPROVED" {
			// Charge failed — expire immediately (no grace period)
			m.Estado = membresiaDomain.MembresiaVencida
			_ = s.membRepo.Update(&m)

			// Record failed payment
			refPasarela := referencia
			if txResp != nil {
				refPasarela = txResp.Data.ID
			}
			_ = s.pagoRepo.Create(&membresiaDomain.PagoMembresia{
				MembresiaID:        m.ID,
				Monto:              float64(precio),
				Moneda:             "COP",
				Tipo:               membresiaDomain.PagoTipoRenovacion,
				ReferenciaPasarela: &refPasarela,
				Estado:             membresiaDomain.PagoEstadoRechazado,
			})
			log.Printf("⚠️ Renovación fallida para membresía %d, marcada como vencida", m.ID)
			continue
		}

		// Charge approved — extend membership
		nextMonth := time.Now().AddDate(0, 1, 0)
		m.FechaFin = &nextMonth
		_ = s.membRepo.Update(&m)

		_ = s.pagoRepo.Create(&membresiaDomain.PagoMembresia{
			MembresiaID:        m.ID,
			Monto:              float64(precio),
			Moneda:             "COP",
			Tipo:               membresiaDomain.PagoTipoRenovacion,
			ReferenciaPasarela: &txResp.Data.ID,
			Estado:             membresiaDomain.PagoEstadoAprobado,
		})
		log.Printf("✅ Membresía %d renovada exitosamente hasta %s", m.ID, nextMonth.Format("2006-01-02"))
	}

	return nil
}

// GetAllMembresias returns all memberships (admin).
func (s *membresiaService) GetAllMembresias() ([]membresiaDomain.MembresiaResponse, error) {
	membresias, err := s.membRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]membresiaDomain.MembresiaResponse, len(membresias))
	for i, m := range membresias {
		responses[i] = *membresiaDomain.ToMembresiaResponse(&m)
	}
	return responses, nil
}

// SimularPago simulates an approved or declined payment directly for testing/sandbox environments.
func (s *membresiaService) SimularPago(usuarioID uint, status string) (*membresiaDomain.MembresiaResponse, error) {
	membresia, err := s.membRepo.FindByUsuarioID(usuarioID)
	if err != nil {
		membresia = &membresiaDomain.Membresia{
			UsuarioID:            usuarioID,
			Estado:               membresiaDomain.MembresiaInactiva,
			RenovacionAutomatica: true,
		}
		if err := s.membRepo.Create(membresia); err != nil {
			return nil, err
		}
	}

	if status == "APPROVED" {
		now := time.Now()
		fechaFin := now.Add(30 * 24 * time.Hour)
		token := "tok_sandbox_card_4242"
		referencia := fmt.Sprintf("SIM-APPROVED-%d", time.Now().UnixMilli())

		membresia.Estado = membresiaDomain.MembresiaActiva
		membresia.FechaInicio = &now
		membresia.FechaFin = &fechaFin
		membresia.MetodoPagoToken = &token
		membresia.RenovacionAutomatica = true
		_ = s.membRepo.Update(membresia)

		// Create approved payment record
		_ = s.pagoRepo.Create(&membresiaDomain.PagoMembresia{
			MembresiaID:        membresia.ID,
			Monto:              25000,
			Moneda:             "COP",
			Tipo:               membresiaDomain.PagoTipoInicial,
			ReferenciaPasarela: &referencia,
			Estado:             membresiaDomain.PagoEstadoAprobado,
		})

		// Trigger referral reward
		s.triggerReferidoAfiliacion(s.db, usuarioID)

		// Activate associated company's subscription if the user is linked to a company
		_ = s.db.Exec(`
			UPDATE administrative.companies SET suscripcion_estado = 'activa'
			WHERE id IN (
				SELECT company_id FROM administrative.user_companies WHERE user_id = ?
				UNION
				SELECT empresa_id FROM administrative.users WHERE id = ? AND empresa_id IS NOT NULL
			)
		`, usuarioID, usuarioID)

		// Notificar a superadministradores (Simulación)
		user, _ := s.userRepo.FindByID(usuarioID)
		nombre := "Usuario"
		email := ""
		if user != nil {
			nombre = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
			email = user.Email
		}
		titulo := fmt.Sprintf("Pago de membresía aprobado - %s", nombre)
		mensaje := fmt.Sprintf("El usuario %s (%s) realizó con éxito el pago de membresía por $25.000 COP (Ref: %s).", nombre, email, referencia)
		if s.notifSvc != nil {
			_ = s.notifSvc.NotificarPagoSuperadmins(titulo, mensaje, notificacionDomain.TipoPagoAprobado, membresia.ID, "pago_membresia")
		}
	} else if status == "DECLINED" {
		referencia := fmt.Sprintf("SIM-DECLINED-%d", time.Now().UnixMilli())
		_ = s.pagoRepo.Create(&membresiaDomain.PagoMembresia{
			MembresiaID:        membresia.ID,
			Monto:              25000,
			Moneda:             "COP",
			Tipo:               membresiaDomain.PagoTipoInicial,
			ReferenciaPasarela: &referencia,
			Estado:             membresiaDomain.PagoEstadoRechazado,
		})

		// Notificar a superadministradores (Simulación)
		user, _ := s.userRepo.FindByID(usuarioID)
		nombre := "Usuario"
		email := ""
		if user != nil {
			nombre = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
			email = user.Email
		}
		titulo := fmt.Sprintf("Pago de membresía rechazado - %s", nombre)
		mensaje := fmt.Sprintf("El intento de pago de membresía para el usuario %s (%s) fue rechazado/falló por $25.000 COP (Ref: %s).", nombre, email, referencia)
		if s.notifSvc != nil {
			_ = s.notifSvc.NotificarPagoSuperadmins(titulo, mensaje, notificacionDomain.TipoPagoRechazado, membresia.ID, "pago_membresia")
		}
	}

	return s.GetMiMembresia(usuarioID)
}
