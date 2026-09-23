package email

import (
	"errors"
	"fmt"
	"log"

	"github.com/resendlabs/resend-go"
)

// EmailService define la interfaz para el envío de correos electrónicos.
type EmailService interface {
	Send(to, subject, htmlBody string) error
	SendActivacionCuenta(to, nombre, token string) error
	SendRecuperacionPassword(to, nombre, codigo string) error
	SendNotificacionPagoSuperadmin(to, adminNombre, usuarioNombre, usuarioEmail, monto, referencia, estado, fecha string) error
}

type resendEmailService struct {
	client      *resend.Client
	fromEmail   string
	frontendURL string
}

// NewEmailService crea una nueva instancia de EmailService utilizando Resend.
func NewEmailService(apiKey, fromEmail, frontendURL string) EmailService {
	if apiKey == "" {
		log.Println("[EmailService] ADVERTENCIA: RESEND_API_KEY no configurada. Los correos no se enviarán.")
	}
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	client := resend.NewClient(apiKey)

	return &resendEmailService{
		client:      client,
		fromEmail:   fromEmail,
		frontendURL: frontendURL,
	}
}

// Send envía un correo electrónico en formato HTML a través de Resend.
func (s *resendEmailService) Send(to, subject, htmlBody string) error {
	if s.client == nil {
		return errors.New("cliente de correo no inicializado")
	}

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("[EmailService] Error enviando correo a %s: %v\n", to, err)
		return fmt.Errorf("error al enviar correo electrónico: %w", err)
	}

	log.Printf("[EmailService] Correo enviado exitosamente a %s (ID: %s)\n", to, sent.Id)
	return nil
}

// SendActivacionCuenta envía el correo de activación para registro asistido con el token de activación.
func (s *resendEmailService) SendActivacionCuenta(to, nombre, token string) error {
	enlace := fmt.Sprintf("%s/activar-cuenta?token=%s", s.frontendURL, token)
	body := PlantillaActivacionCuenta(nombre, enlace)
	return s.Send(to, "Activa tu cuenta en Conexiate", body)
}

// SendRecuperacionPassword envía el correo para restablecer contraseña con el código de 6 dígitos.
func (s *resendEmailService) SendRecuperacionPassword(to, nombre, codigo string) error {
	body := PlantillaRecuperacionPassword(nombre, codigo)
	return s.Send(to, "Tu código de recuperación - Conexiate", body)
}

// SendNotificacionPagoSuperadmin envía un correo al superadministrador notificando el pago de una membresía.
func (s *resendEmailService) SendNotificacionPagoSuperadmin(to, adminNombre, usuarioNombre, usuarioEmail, monto, referencia, estado, fecha string) error {
	asunto := fmt.Sprintf("🔔 Notificación de Pago de Membresía (%s) - Conexiate", estado)
	enlaceAdmin := fmt.Sprintf("%s/dashboard", s.frontendURL)
	body := PlantillaNotificacionPagoSuperadmin(adminNombre, usuarioNombre, usuarioEmail, monto, referencia, estado, fecha, enlaceAdmin)
	return s.Send(to, asunto, body)
}
