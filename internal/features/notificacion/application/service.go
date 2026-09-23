package application

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"multicliente-backend/internal/features/notificacion/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/email"
)

type notificacionService struct {
	repo     domain.NotificacionRepository
	db       *gorm.DB
	emailSvc email.EmailService
}

// NewNotificacionService crea una instancia del servicio de notificaciones con soporte de correo.
func NewNotificacionService(repo domain.NotificacionRepository, db *gorm.DB, emailSvc email.EmailService) domain.NotificacionService {
	return &notificacionService{
		repo:     repo,
		db:       db,
		emailSvc: emailSvc,
	}
}

func (s *notificacionService) GetNotificaciones(usuarioID uint, limit, offset int) ([]domain.Notificacion, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.FindByUsuarioID(usuarioID, limit, offset)
}

func (s *notificacionService) GetCountNoLeidas(usuarioID uint) (int64, error) {
	return s.repo.CountNoLeidas(usuarioID)
}

func (s *notificacionService) MarcarLeida(id uint, usuarioID uint) error {
	return s.repo.MarcarLeida(id, usuarioID)
}

func (s *notificacionService) MarcarTodasLeidas(usuarioID uint) error {
	return s.repo.MarcarTodasLeidas(usuarioID)
}

// NotificarPagoSuperadmins crea una notificación para cada superadministrador activo del sistema.
func (s *notificacionService) NotificarPagoSuperadmins(titulo, mensaje, tipo string, pagoID uint, referenciaTipo string) error {
	var superadminUsers []userDomain.User
	err := s.db.Raw(`
		SELECT u.* 
		FROM administrative.users u 
		JOIN administrative.roles r ON r.id = u.role_id 
		WHERE r.code IN ('superadmin', 'admin') AND u.is_active = true
	`).Scan(&superadminUsers).Error

	if err != nil {
		log.Printf("[NotificacionService] Error buscando administradores: %v\n", err)
		return err
	}

	if len(superadminUsers) == 0 {
		log.Println("[NotificacionService] No se encontraron superadministradores activos para notificar.")
		return nil
	}

	refID := pagoID
	refTipo := referenciaTipo

	// Extraer detalles de la transacción para el correo electrónico
	estadoStr := "aprobado"
	if tipo == domain.TipoPagoRechazado {
		estadoStr = "rechazado"
	}
	montoStr := "25.000"
	usuarioNombre := "Afiliado"
	usuarioEmail := ""
	referenciaStr := ""
	fechaStr := time.Now().Format("02/01/2006 15:04")

	if pagoID > 0 {
		var pagoInfo struct {
			Monto              float64
			ReferenciaPasarela string
			UsuarioID          uint
			FechaPago          time.Time
		}
		if err := s.db.Table("administrative.pago_membresia p").
			Select("p.monto, COALESCE(p.referencia_pasarela, '') as referencia_pasarela, m.usuario_id, p.fecha_pago").
			Joins("JOIN administrative.membresia m ON m.id = p.membresia_id").
			Where("p.id = ?", pagoID).
			Scan(&pagoInfo).Error; err == nil && pagoInfo.UsuarioID > 0 {
			montoStr = fmt.Sprintf("%.0f", pagoInfo.Monto)
			referenciaStr = pagoInfo.ReferenciaPasarela
			if !pagoInfo.FechaPago.IsZero() {
				fechaStr = pagoInfo.FechaPago.Format("02/01/2006 15:04")
			}
			var u userDomain.User
			if err := s.db.First(&u, pagoInfo.UsuarioID).Error; err == nil {
				usuarioNombre = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
				usuarioEmail = u.Email
			}
		}
	}

	var notificaciones []domain.Notificacion
	for _, admin := range superadminUsers {
		notificaciones = append(notificaciones, domain.Notificacion{
			UsuarioID:      admin.ID,
			Tipo:           tipo,
			Titulo:         titulo,
			Mensaje:        mensaje,
			Leido:          false,
			ReferenciaID:   &refID,
			ReferenciaTipo: &refTipo,
		})

		// Enviar correo electrónico al superadministrador si emailSvc está configurado
		if s.emailSvc != nil && admin.Email != "" {
			adminEmail := admin.Email
			adminName := admin.FirstName
			if adminName == "" {
				adminName = "Administrador"
			}
			go func(to, aName, uName, uEmail, m, ref, est, f string) {
				if err := s.emailSvc.SendNotificacionPagoSuperadmin(to, aName, uName, uEmail, m, ref, est, f); err != nil {
					log.Printf("[NotificacionService] Error enviando correo de pago a %s: %v\n", to, err)
				}
			}(adminEmail, adminName, usuarioNombre, usuarioEmail, montoStr, referenciaStr, estadoStr, fechaStr)
		}
	}

	if err := s.repo.CreateBatch(notificaciones); err != nil {
		log.Printf("[NotificacionService] Error creando notificaciones en lote: %v\n", err)
		return err
	}

	log.Printf("[NotificacionService] Notificación enviada exitosamente a %d superadministradores (Pago ID: %d)\n", len(superadminUsers), pagoID)
	return nil
}
