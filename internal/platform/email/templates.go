package email

import (
	"fmt"
	"html"
)

// PlantillaActivacionCuenta genera el HTML responsivo para el correo de activación de cuenta (registro asistido).
func PlantillaActivacionCuenta(nombre, enlace string) string {
	nombreSeguro := html.EscapeString(nombre)
	enlaceSeguro := html.EscapeString(enlace)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Activa tu cuenta en Conexiate</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      background-color: #f4f6f9;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      color: #1e293b;
      -webkit-font-smoothing: antialiased;
    }
    .container {
      max-width: 580px;
      margin: 40px auto;
      background: #ffffff;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
      border: 1px solid #e2e8f0;
    }
    .header {
      background: linear-gradient(135deg, #1e3a8a 0%%, #3b82f6 100%%);
      padding: 32px 24px;
      text-align: center;
    }
    .header h1 {
      margin: 0;
      color: #ffffff;
      font-size: 24px;
      font-weight: 700;
      letter-spacing: -0.5px;
    }
    .content {
      padding: 36px 32px;
    }
    .greeting {
      font-size: 18px;
      font-weight: 600;
      color: #0f172a;
      margin-bottom: 16px;
    }
    .paragraph {
      font-size: 15px;
      line-height: 1.6;
      color: #475569;
      margin-bottom: 24px;
    }
    .btn-container {
      text-align: center;
      margin: 32px 0;
    }
    .btn {
      display: inline-block;
      background: #2563eb;
      color: #ffffff !important;
      text-decoration: none;
      padding: 14px 32px;
      border-radius: 8px;
      font-size: 15px;
      font-weight: 600;
      letter-spacing: 0.2px;
      box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
    }
    .fallback-box {
      background: #f8fafc;
      border-radius: 8px;
      padding: 16px;
      border: 1px dashed #cbd5e1;
      margin-top: 24px;
    }
    .fallback-box p {
      margin: 0 0 8px 0;
      font-size: 13px;
      color: #64748b;
    }
    .fallback-box a {
      color: #2563eb;
      font-size: 13px;
      word-break: break-all;
    }
    .note {
      font-size: 13px;
      color: #94a3b8;
      margin-top: 24px;
      line-height: 1.5;
    }
    .footer {
      background: #f8fafc;
      padding: 20px 32px;
      text-align: center;
      border-top: 1px solid #e2e8f0;
      font-size: 12px;
      color: #94a3b8;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>CONEXIATE</h1>
    </div>
    <div class="content">
      <div class="greeting">¡Hola, %s! 👋</div>
      <p class="paragraph">
        Tu cuenta en la plataforma de beneficios <strong>Conexiate</strong> ha sido creada satisfactoriamente.
      </p>
      <p class="paragraph">
        Para comenzar a disfrutar de todos tus beneficios y servicios, por favor haz clic en el siguiente botón para activar tu cuenta y establecer tu contraseña de acceso:
      </p>
      <div class="btn-container">
        <a href="%s" class="btn" target="_blank">Activar mi cuenta</a>
      </div>
      <div class="fallback-box">
        <p>Si el botón no funciona, copia y pega este enlace en tu navegador:</p>
        <a href="%s" target="_blank">%s</a>
      </div>
      <p class="note">
        ⏰ <strong>Importante:</strong> Este enlace de activación es de uso único y permanecerá activo durante las próximas 24 horas por motivos de seguridad.
      </p>
    </div>
    <div class="footer">
      <p style="margin: 0;">© Conexiate. Todos los derechos reservados.</p>
    </div>
  </div>
</body>
</html>`, nombreSeguro, enlaceSeguro, enlaceSeguro, enlaceSeguro)
}

// PlantillaRecuperacionPassword genera el HTML responsivo para el correo de restablecimiento de contraseña con código numérico de 6 dígitos.
func PlantillaRecuperacionPassword(nombre, codigo string) string {
	nombreSeguro := html.EscapeString(nombre)
	codigoSeguro := html.EscapeString(codigo)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Código de recuperación - Conexiate</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      background-color: #f4f6f9;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      color: #1e293b;
      -webkit-font-smoothing: antialiased;
    }
    .container {
      max-width: 580px;
      margin: 40px auto;
      background: #ffffff;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
      border: 1px solid #e2e8f0;
    }
    .header {
      background: linear-gradient(135deg, #1e3a8a 0%%, #3b82f6 100%%);
      padding: 32px 24px;
      text-align: center;
    }
    .header h1 {
      margin: 0;
      color: #ffffff;
      font-size: 24px;
      font-weight: 700;
      letter-spacing: -0.5px;
    }
    .content {
      padding: 36px 32px;
    }
    .greeting {
      font-size: 18px;
      font-weight: 600;
      color: #0f172a;
      margin-bottom: 16px;
    }
    .paragraph {
      font-size: 15px;
      line-height: 1.6;
      color: #475569;
      margin-bottom: 24px;
    }
    .code-container {
      text-align: center;
      margin: 28px 0;
      background: #f8fafc;
      border: 2px dashed #93c5fd;
      border-radius: 12px;
      padding: 24px 20px;
    }
    .code-label {
      margin: 0 0 10px 0;
      font-size: 13px;
      font-weight: 600;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .code-box {
      display: inline-block;
      font-family: 'Courier New', Courier, monospace;
      font-size: 38px;
      font-weight: 800;
      letter-spacing: 12px;
      color: #1e40af;
      background: #ffffff;
      padding: 10px 24px;
      border-radius: 8px;
      border: 1px solid #bfdbfe;
      box-shadow: 0 2px 4px rgba(0,0,0,0.05);
    }
    .code-expiry {
      margin: 12px 0 0 0;
      font-size: 13px;
      color: #b45309;
      font-weight: 500;
    }
    .security-box {
      background: #fffbeb;
      border-left: 4px solid #f59e0b;
      padding: 14px 16px;
      border-radius: 4px;
      margin-top: 24px;
    }
    .security-box p {
      margin: 0;
      font-size: 13px;
      color: #92400e;
      line-height: 1.5;
    }
    .footer {
      background: #f8fafc;
      padding: 20px 32px;
      text-align: center;
      border-top: 1px solid #e2e8f0;
      font-size: 12px;
      color: #94a3b8;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>CONEXIATE</h1>
    </div>
    <div class="content">
      <div class="greeting">¡Hola, %s! 👋</div>
      <p class="paragraph">
        Hemos recibido una solicitud para restablecer la contraseña de acceso a tu cuenta en <strong>Conexiate</strong>.
      </p>
      <p class="paragraph">
        Ingresa el siguiente código de 6 dígitos en la aplicación para verificar tu identidad:
      </p>
      <div class="code-container">
        <p class="code-label">Tu código de verificación es:</p>
        <div class="code-box">%s</div>
        <p class="code-expiry">⏱️ Este código expira en <strong>10 minutos</strong>.</p>
      </div>
      <div class="security-box">
        <p>🛡️ <strong>Aviso de seguridad:</strong> Si tú no solicitaste este cambio, puedes ignorar este mensaje con tranquilidad. Tu contraseña permanecerá intacta y nadie podrá acceder sin tu autorización.</p>
      </div>
    </div>
    <div class="footer">
      <p style="margin: 0;">© Conexiate. Todos los derechos reservados.</p>
    </div>
  </div>
 </body>
 </html>`, nombreSeguro, codigoSeguro)
}

// PlantillaNotificacionPagoSuperadmin genera el HTML para notificar a los superadministradores sobre pagos de membresía.
func PlantillaNotificacionPagoSuperadmin(adminNombre, usuarioNombre, usuarioEmail, monto, referencia, estado, fecha, enlaceAdmin string) string {
	adminSeguro := html.EscapeString(adminNombre)
	usuarioSeguro := html.EscapeString(usuarioNombre)
	emailSeguro := html.EscapeString(usuarioEmail)
	montoSeguro := html.EscapeString(monto)
	refSeguro := html.EscapeString(referencia)
	fechaSegura := html.EscapeString(fecha)
	enlaceSeguro := html.EscapeString(enlaceAdmin)

	isAprobado := estado == "aprobado" || estado == "APPROVED"
	estadoColor := "#10b981"
	estadoTexto := "PAGO APROBADO"
	if !isAprobado {
		estadoColor = "#ef4444"
		estadoTexto = "PAGO RECHAZADO / FALLIDO"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Notificación de Pago - Conexiate</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      background-color: #f4f6f9;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      color: #1e293b;
      -webkit-font-smoothing: antialiased;
    }
    .container {
      max-width: 580px;
      margin: 40px auto;
      background: #ffffff;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
      border: 1px solid #e2e8f0;
    }
    .header {
      background: linear-gradient(135deg, #1e3a8a 0%%, #0f172a 100%%);
      padding: 28px 24px;
      text-align: center;
    }
    .header h1 {
      margin: 0;
      color: #ffffff;
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 1px;
    }
    .header-badge {
      display: inline-block;
      margin-top: 8px;
      padding: 4px 12px;
      background: rgba(255, 255, 255, 0.15);
      border-radius: 20px;
      color: #93c5fd;
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .content {
      padding: 32px 28px;
    }
    .greeting {
      font-size: 17px;
      font-weight: 600;
      color: #0f172a;
      margin-bottom: 12px;
    }
    .paragraph {
      font-size: 14px;
      line-height: 1.6;
      color: #475569;
      margin-bottom: 20px;
    }
    .status-badge {
      display: inline-block;
      padding: 6px 14px;
      background: %s;
      color: #ffffff;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.5px;
      margin-bottom: 20px;
    }
    .details-box {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 10px;
      padding: 18px 20px;
      margin-bottom: 24px;
    }
    .detail-row {
      display: flex;
      justify-content: space-between;
      padding: 8px 0;
      border-bottom: 1px solid #edf2f7;
      font-size: 13px;
    }
    .detail-row:last-child {
      border-bottom: none;
    }
    .detail-label {
      color: #64748b;
      font-weight: 500;
    }
    .detail-value {
      color: #0f172a;
      font-weight: 600;
      text-align: right;
    }
    .monto-highlight {
      color: #10b981;
      font-size: 16px;
      font-weight: 700;
    }
    .btn-container {
      text-align: center;
      margin: 28px 0 10px 0;
    }
    .btn {
      display: inline-block;
      background: #2563eb;
      color: #ffffff !important;
      text-decoration: none;
      padding: 12px 28px;
      border-radius: 8px;
      font-size: 14px;
      font-weight: 600;
      box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
    }
    .footer {
      background: #f8fafc;
      padding: 18px 28px;
      text-align: center;
      border-top: 1px solid #e2e8f0;
      font-size: 12px;
      color: #94a3b8;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>CONEXIATE</h1>
      <div class="header-badge">Panel Administrativo</div>
    </div>
    <div class="content">
      <div class="greeting">Hola, %s 👋</div>
      <p class="paragraph">
        Te notificamos que se ha registrado un nuevo evento de pago de membresía en la plataforma a través de la pasarela Wompi:
      </p>
      <div>
        <span class="status-badge">%s</span>
      </div>
      <div class="details-box">
        <table style="width: 100%%; border-collapse: collapse;">
          <tr style="border-bottom: 1px solid #edf2f7;">
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Afiliado:</td>
            <td style="padding: 8px 0; color: #0f172a; font-weight: 600; font-size: 13px; text-align: right;">%s</td>
          </tr>
          <tr style="border-bottom: 1px solid #edf2f7;">
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Correo:</td>
            <td style="padding: 8px 0; color: #0f172a; font-weight: 500; font-size: 13px; text-align: right;">%s</td>
          </tr>
          <tr style="border-bottom: 1px solid #edf2f7;">
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Concepto:</td>
            <td style="padding: 8px 0; color: #0f172a; font-weight: 600; font-size: 13px; text-align: right;">Membresía Mensual</td>
          </tr>
          <tr style="border-bottom: 1px solid #edf2f7;">
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Monto:</td>
            <td style="padding: 8px 0; color: #10b981; font-weight: 700; font-size: 15px; text-align: right;">$%s COP</td>
          </tr>
          <tr style="border-bottom: 1px solid #edf2f7;">
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Referencia Wompi:</td>
            <td style="padding: 8px 0; color: #334155; font-family: monospace; font-size: 12px; text-align: right;">%s</td>
          </tr>
          <tr>
            <td style="padding: 8px 0; color: #64748b; font-size: 13px;">Fecha:</td>
            <td style="padding: 8px 0; color: #64748b; font-size: 13px; text-align: right;">%s</td>
          </tr>
        </table>
      </div>
      <div class="btn-container">
        <a href="%s" class="btn">Abrir Panel de Control</a>
      </div>
    </div>
    <div class="footer">
      <p style="margin: 0;">© Conexiate. Este es un correo automático de control para superadministradores.</p>
    </div>
  </div>
</body>
</html>`, estadoColor, adminSeguro, estadoTexto, usuarioSeguro, emailSeguro, montoSeguro, refSeguro, fechaSegura, enlaceSeguro)
}
