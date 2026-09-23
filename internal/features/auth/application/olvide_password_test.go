package application_test

import (
	"errors"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	authApplication "multicliente-backend/internal/features/auth/application"
	authDomain "multicliente-backend/internal/features/auth/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	userInfrastructure "multicliente-backend/internal/features/user/infrastructure"
	"multicliente-backend/internal/platform/config"
	"multicliente-backend/internal/platform/database"
	"multicliente-backend/internal/platform/email"
)

type mockUserRepo struct {
	userDomain.UserRepository
	findByEmailFunc func(email string) (*userDomain.User, error)
}

func (m *mockUserRepo) FindByEmail(email string) (*userDomain.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(email)
	}
	return nil, errors.New("record not found")
}

type mockEmailService struct {
	email.EmailService
	mu       sync.Mutex
	lastTo   string
	lastCode string
}

func (m *mockEmailService) SendRecuperacionPassword(to, name, codigo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastTo = to
	m.lastCode = codigo
	return nil
}

func (m *mockEmailService) LastCode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCode
}

func TestOlvidePassword_UsuarioNoExiste(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFunc: func(email string) (*userDomain.User, error) {
			return nil, errors.New("record not found")
		},
	}

	svc := authApplication.NewAuthService(repo, nil, nil, nil, nil, "secret", "24")
	err := svc.OlvidePassword(&authDomain.OlvidePasswordRequest{
		Email: "noexiste@correo.com",
	})

	if err == nil {
		t.Fatalf("se esperaba un error al solicitar recuperación para un usuario inexistente, pero fue nil")
	}

	expectedMsg := "no se encontró ningún usuario registrado con este correo electrónico"
	if err.Error() != expectedMsg {
		t.Fatalf("mensaje de error incorrecto: se esperaba '%s', se obtuvo '%s'", expectedMsg, err.Error())
	}
}

func TestOlvidePassword_UsuarioInactivo(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFunc: func(email string) (*userDomain.User, error) {
			return &userDomain.User{
				Email:    "inactivo@correo.com",
				IsActive: false,
			}, nil
		},
	}

	svc := authApplication.NewAuthService(repo, nil, nil, nil, nil, "secret", "24")
	err := svc.OlvidePassword(&authDomain.OlvidePasswordRequest{
		Email: "inactivo@correo.com",
	})

	if err == nil {
		t.Fatalf("se esperaba un error al solicitar recuperación para un usuario inactivo, pero fue nil")
	}

	expectedMsg := "la cuenta de usuario se encuentra inactiva"
	if err.Error() != expectedMsg {
		t.Fatalf("mensaje de error incorrecto: se esperaba '%s', se obtuvo '%s'", expectedMsg, err.Error())
	}
}

func TestOlvidePassword_FlujoCompletoCodigoYReset(t *testing.T) {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("Omitiendo prueba de integración de base de datos: %v", err)
	}

	_ = db.Exec(`
		CREATE TABLE IF NOT EXISTS administrative.password_reset_code (
			id BIGSERIAL PRIMARY KEY,
			usuario_id BIGINT NOT NULL REFERENCES administrative.users(id),
			codigo_hash VARCHAR(255) NOT NULL,
			intentos_fallidos INT NOT NULL DEFAULT 0,
			usado BOOLEAN NOT NULL DEFAULT false,
			expira_at TIMESTAMPTZ NOT NULL,
			create_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_password_reset_code_usuario ON administrative.password_reset_code(usuario_id);
	`).Error

	userRepo := userInfrastructure.NewUserRepository(db)
	mockEmail := &mockEmailService{}
	jwtSecret := "test_jwt_secret_key_12345"

	svc := authApplication.NewAuthService(userRepo, nil, nil, db, mockEmail, jwtSecret, "24")

	// Crear usuario de prueba temporal
	testEmail := "test_otp_user_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@test.com"
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	testUser := &userDomain.User{
		Email:     testEmail,
		Password:  string(hashedPass),
		FirstName: "Test",
		LastName:  "OTP",
		IsActive:  true,
	}
	if err := db.Create(testUser).Error; err != nil {
		t.Fatalf("Error creando usuario de prueba: %v", err)
	}
	defer func() {
		_ = db.Where("usuario_id = ?", testUser.ID).Delete(&authDomain.PasswordResetCode{})
		_ = db.Delete(testUser)
	}()

	// 1. Solicitar recuperación
	err = svc.OlvidePassword(&authDomain.OlvidePasswordRequest{
		Email: testEmail,
	})
	if err != nil {
		t.Fatalf("OlvidePassword falló inesperadamente: %v", err)
	}

	// Esperar goroutine de correo
	time.Sleep(150 * time.Millisecond)
	codigoGenerado := mockEmail.LastCode()
	if len(codigoGenerado) != 6 {
		t.Fatalf("Se esperaba un código de 6 dígitos, se obtuvo '%s'", codigoGenerado)
	}
	if !regexp.MustCompile(`^\d{6}$`).MatchString(codigoGenerado) {
		t.Fatalf("El código generado debe ser estrictamente numérico de 6 dígitos: '%s'", codigoGenerado)
	}

	// 2. Verificar con código erróneo
	_, err = svc.VerificarCodigo(&authDomain.VerificarCodigoRequest{
		Email:  testEmail,
		Codigo: "000000",
	})
	if err == nil {
		t.Fatalf("Se esperaba error con código incorrecto, pero fue exitoso")
	}

	// 3. Verificar con código correcto
	resp, err := svc.VerificarCodigo(&authDomain.VerificarCodigoRequest{
		Email:  testEmail,
		Codigo: codigoGenerado,
	})
	if err != nil {
		t.Fatalf("VerificarCodigo falló con código correcto: %v", err)
	}
	if resp.ResetToken == "" {
		t.Fatalf("Se esperaba un ResetToken no vacío en la respuesta")
	}

	// 4. Restablecer contraseña con el reset_token emitido
	newPassword := "NuevaPasswordSegura2026!"
	err = svc.ResetPassword(&authDomain.ResetPasswordRequest{
		Token:    resp.ResetToken,
		Password: newPassword,
	})
	if err != nil {
		t.Fatalf("ResetPassword falló con el reset_token: %v", err)
	}

	// 5. Verificar que la contraseña en BD fue actualizada
	updatedUser, err := userRepo.FindByID(testUser.ID)
	if err != nil {
		t.Fatalf("Error obteniendo usuario actualizado: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword)); err != nil {
		t.Fatalf("La contraseña en base de datos no coincide con la nueva contraseña")
	}

	// 6. Verificar que el código quedó marcado como usado
	var codeRecord authDomain.PasswordResetCode
	if err := db.Where("usuario_id = ?", testUser.ID).First(&codeRecord).Error; err == nil {
		if !codeRecord.Usado {
			t.Fatalf("Se esperaba que el registro password_reset_code tuviera usado = true")
		}
	}
}
