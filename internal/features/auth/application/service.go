package application

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authDomain "multicliente-backend/internal/features/auth/domain"
	companyDomain "multicliente-backend/internal/features/company/domain"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/email"
	"multicliente-backend/internal/platform/validation"
)

type authService struct {
	userRepo    userDomain.UserRepository
	referidoSvc referidoDomain.ReferidoService
	companyRepo companyDomain.CompanyRepository
	db          *gorm.DB
	emailSvc    email.EmailService
	jwtSecret   string
	jwtExpHours int
}

// NewAuthService creates a new AuthService.
// It depends on the user repository, referido service, company repository, DB, and email service
// for handling referral logic, company code linking, dynamic role assignment, and account emails.
func NewAuthService(
	userRepo userDomain.UserRepository,
	referidoSvc referidoDomain.ReferidoService,
	companyRepo companyDomain.CompanyRepository,
	db *gorm.DB,
	emailSvc email.EmailService,
	jwtSecret string,
	jwtExpHours string,
) authDomain.AuthService {
	hours, err := strconv.Atoi(jwtExpHours)
	if err != nil || hours <= 0 {
		hours = 24
	}
	return &authService{
		userRepo:    userRepo,
		referidoSvc: referidoSvc,
		companyRepo: companyRepo,
		db:          db,
		emailSvc:    emailSvc,
		jwtSecret:   jwtSecret,
		jwtExpHours: hours,
	}
}

func (s *authService) Login(req *authDomain.LoginRequest) (*authDomain.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Check if user account is temporarily locked due to failed attempts
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		remainingMinutes := int(time.Until(*user.LockedUntil).Minutes()) + 1
		return nil, fmt.Errorf("cuenta_bloqueada:%d", remainingMinutes)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		user.FailedAttempts++
		if user.FailedAttempts >= 5 {
			lockTime := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockTime
			_ = s.userRepo.Update(user)
			return nil, fmt.Errorf("cuenta_bloqueada:15")
		}
		_ = s.userRepo.Update(user)
		remainingAttempts := 5 - user.FailedAttempts
		return nil, fmt.Errorf("credenciales_invalidas:%d", remainingAttempts)
	}

	// Reset failed attempts on successful password verification
	if user.FailedAttempts > 0 || user.LockedUntil != nil {
		user.FailedAttempts = 0
		user.LockedUntil = nil
		_ = s.userRepo.Update(user)
	}

	// Extract role details
	roleCode := ""
	var roleIDVal *uint
	if user.Role != nil {
		roleCode = user.Role.Code
		roleIDVal = &user.Role.ID
	}

	// Determine token expiration duration
	expirationDuration := time.Hour * time.Duration(s.jwtExpHours)
	if user.Role != nil {
		roleDuration := time.Duration(user.Role.SessionDays)*24*time.Hour +
			time.Duration(user.Role.SessionHours)*time.Hour +
			time.Duration(user.Role.SessionMinutes)*time.Minute
		if roleDuration > 0 {
			expirationDuration = roleDuration
		}
	}

	// Generate JWT token
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"role":     roleCode,
		"role_id":  roleIDVal,
		"iat":      now.Unix(),
		"exp":      now.Add(expirationDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &authDomain.LoginResponse{
		Token: tokenString,
		User: authDomain.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleID:    roleIDVal,
			RoleCode:  roleCode,
		},
		SessionDurationSeconds: int(expirationDuration.Seconds()),
	}, nil
}

// Register creates a new user account (public endpoint).
// Order of resolution:
// 1. Validate ref_code -> determine refCodeValido
// 2. Validate codigo_empresa -> determine codigoEmpresaValido
// 3. Determine role_id dynamically (user or user_member)
// 4. Create user
// 5. Handle referido linking if refCodeValido
// 6. Handle user_companies linking if codigoEmpresaValido
func (s *authService) Register(req *authDomain.RegisterRequest) (*authDomain.RegisterResponse, error) {
	// 1. Validate that email doesn't already exist
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("el correo electrónico ya está registrado")
	}

	// 2. Validate password policies
	nombreCompleto := req.FirstName + " " + req.LastName
	if err := validation.ValidarPassword(req.Password, req.Email, nombreCompleto); err != nil {
		return nil, err
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("falló al procesar la contraseña")
	}

	// 4. Generate unique code_refer for the new user
	codeRefer, err := generateCodeRefer()
	if err != nil {
		return nil, errors.New("falló al generar código de referencia")
	}

	// === Step 1: Validate ref_code ===
	var referente *userDomain.User
	refCodeValido := false
	if req.RefCode != nil && *req.RefCode != "" {
		ref, refErr := s.userRepo.FindByCodeRefer(*req.RefCode)
		if refErr == nil && ref != nil {
			referente = ref
			refCodeValido = true
		}
		// If invalid, silently ignore
	}

	// === Step 2: Validate codigo_empresa ===
	var companyForLink *companyDomain.Company
	codigoEmpresaValido := false
	if req.CodigoEmpresa != nil && *req.CodigoEmpresa != "" {
		company, compErr := s.companyRepo.FindByCodigoEmpresa(*req.CodigoEmpresa)
		if compErr == nil && company != nil {
			companyForLink = company
			codigoEmpresaValido = true
		}
		// If invalid, silently ignore
	}

	// === Step 3: Determine role dynamically ===
	var codeRol string
	if refCodeValido || codigoEmpresaValido {
		codeRol = "user_member"
	} else {
		codeRol = "user"
	}

	var rol roleDomain.Role
	if err := s.db.Where("code = ?", codeRol).First(&rol).Error; err != nil {
		return nil, errors.New("no se encontró el rol '" + codeRol + "' en el sistema")
	}

	// === Step 4: Create the user ===
	user := &userDomain.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		RoleID:    &rol.ID,
		CodeRefer: codeRefer,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// === Step 5: Handle referral logic if ref_code was valid ===
	referidoCreated := false
	if refCodeValido && referente != nil && referente.ID != user.ID {
		// Try to convert an existing pending referral, or create + convert
		convertedRef, convertErr := s.referidoSvc.ConvertirReferido(req.Email, user.ID)
		if convertErr != nil {
			// No pending invitation exists — create one, then convert
			nombre := req.FirstName + " " + req.LastName
			createReq := &referidoDomain.CreateReferidoRequest{
				EmailReferido:  req.Email,
				NombreReferido: &nombre,
			}
			_, createErr := s.referidoSvc.CreateReferido(createReq, referente.ID)
			if createErr == nil {
				// Convert to 'registrado'
				convertedRef, _ = s.referidoSvc.ConvertirReferido(req.Email, user.ID)
			}
		}
		// NOTE: Auto-affiliation REMOVED here. Referido stays in 'registrado'.
		// Transition to 'afiliado' happens in the Wompi webhook when first payment is approved.
		if convertedRef != nil {
			referidoCreated = true
		}
	}

	// === Step 6: Handle user_companies linking if codigo_empresa was valid ===
	if codigoEmpresaValido && companyForLink != nil {
		_ = s.db.Exec(
			"INSERT INTO administrative.user_companies (user_id, company_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			user.ID, companyForLink.ID,
		)
	}

	// 7. Reload user with relations for the response
	fullUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}

	// 8. Generate token for auto-login after registration
	roleCode := ""
	var roleIDVal *uint
	if fullUser.Role != nil {
		roleCode = fullUser.Role.Code
		roleIDVal = &fullUser.Role.ID
	}

	expirationDuration := time.Hour * time.Duration(s.jwtExpHours)
	if fullUser.Role != nil {
		roleDuration := time.Duration(fullUser.Role.SessionDays)*24*time.Hour +
			time.Duration(fullUser.Role.SessionHours)*time.Hour +
			time.Duration(fullUser.Role.SessionMinutes)*time.Minute
		if roleDuration > 0 {
			expirationDuration = roleDuration
		}
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": fullUser.ID,
		"email":   fullUser.Email,
		"role":    roleCode,
		"role_id": roleIDVal,
		"iat":     now.Unix(),
		"exp":     now.Add(expirationDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &authDomain.RegisterResponse{
		Token: tokenString,
		User: authDomain.UserInfo{
			ID:        fullUser.ID,
			Email:     fullUser.Email,
			FirstName: fullUser.FirstName,
			LastName:  fullUser.LastName,
			RoleID:    roleIDVal,
			RoleCode:  roleCode,
		},
		SessionDurationSeconds: int(expirationDuration.Seconds()),
		ReferidoCreated:        referidoCreated,
	}, nil
}

func (s *authService) GetProfile(userID uint) (*userDomain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if len(user.Companies) == 0 {
		var empresaID *uint
		_ = s.db.Table("administrative.users").Select("empresa_id").Where("id = ?", userID).Scan(&empresaID)
		if empresaID != nil && *empresaID > 0 {
			var comp companyDomain.Company
			if err := s.db.Table("administrative.companies").First(&comp, "id = ?", *empresaID).Error; err == nil {
				user.Companies = []companyDomain.Company{comp}
				_ = s.db.Exec("INSERT INTO administrative.user_companies (user_id, company_id) VALUES (?, ?) ON CONFLICT DO NOTHING", userID, *empresaID)
			}
		}
	}

	// Hide password hash
	user.Password = ""
	return user, nil
}

// GetMyCodeRefer returns the authenticated user's own referral code.
func (s *authService) GetMyCodeRefer(userID uint) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("usuario no encontrado")
	}
	if user.CodeRefer == "" {
		code, err := generateCodeRefer()
		if err == nil {
			user.CodeRefer = code
			_ = s.userRepo.Update(user)
		}
	}
	return user.CodeRefer, nil
}

func (s *authService) ChangePassword(userID uint, req *authDomain.ChangePasswordRequest) error {
	if req.CurrentPassword != req.ConfirmCurrentPassword {
		return errors.New("la contraseña actual y la confirmación no coinciden")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("la contraseña actual es incorrecta")
	}

	// Validate new password policies
	nombreCompleto := user.FirstName + " " + user.LastName
	if err := validation.ValidarPassword(req.NewPassword, user.Email, nombreCompleto); err != nil {
		return err
	}

	// Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("falló al procesar la nueva contraseña")
	}

	user.Password = string(hashed)
	return s.userRepo.Update(user)
}

func (s *authService) UpdateProfile(userID uint, req *authDomain.UpdateProfileRequest) (*userDomain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	// Verify email unique check
	if req.Email != user.Email {
		existing, _ := s.userRepo.FindByEmail(req.Email)
		if existing != nil && existing.ID != userID {
			return nil, errors.New("el correo electrónico ya está registrado")
		}
	}

	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.UpdateBy = &userID

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Fetch fully loaded user
	fullUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Hide password hash
	fullUser.Password = ""
	return fullUser, nil
}

// ValidarCodigoEmpresa validates a company code and returns the company name.
func (s *authService) ValidarCodigoEmpresa(codigo string) (*companyDomain.Company, error) {
	company, err := s.companyRepo.FindByCodigoEmpresa(codigo)
	if err != nil {
		return nil, errors.New("código de empresa no encontrado")
	}
	return company, nil
}

// OlvidePassword valida la existencia y estado del usuario, genera un código numérico de 6 dígitos,
// invalida códigos anteriores, guarda el hash en administrative.password_reset_code y envía el correo.
func (s *authService) OlvidePassword(req *authDomain.OlvidePasswordRequest) error {
	emailBuscado := strings.TrimSpace(req.Email)
	user, err := s.userRepo.FindByEmail(emailBuscado)
	if err != nil || user == nil {
		log.Printf("[Auth] OlvidePassword: Solicitud rechazada, usuario no encontrado (%s)\n", emailBuscado)
		return errors.New("no se encontró ningún usuario registrado con este correo electrónico")
	}

	if !user.IsActive {
		log.Printf("[Auth] OlvidePassword: Solicitud rechazada, cuenta inactiva (%s)\n", emailBuscado)
		return errors.New("la cuenta de usuario se encuentra inactiva")
	}

	// 1. Generar código numérico de 6 dígitos
	codigo, err := generateNumericCode()
	if err != nil {
		return errors.New("falló al generar código de recuperación")
	}

	// 2. Hashear el código con bcrypt
	hashedCodigo, err := bcrypt.GenerateFromPassword([]byte(codigo), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("falló al procesar código de verificación")
	}

	// 3. Invalidar cualquier código anterior no usado de este usuario
	if err := s.db.Model(&authDomain.PasswordResetCode{}).
		Where("usuario_id = ? AND usado = false", user.ID).
		Update("usado", true).Error; err != nil {
		log.Printf("[Auth] Advertencia al invalidar códigos previos de %s: %v\n", user.Email, err)
	}

	// 4. Guardar el nuevo código con expiración en 10 minutos
	resetCode := &authDomain.PasswordResetCode{
		UsuarioID:        user.ID,
		CodigoHash:       string(hashedCodigo),
		IntentosFallidos: 0,
		Usado:            false,
		ExpiraAt:         time.Now().Add(10 * time.Minute),
	}

	if err := s.db.Create(resetCode).Error; err != nil {
		log.Printf("[Auth] Error guardando código de recuperación para %s: %v\n", user.Email, err)
		return errors.New("error al procesar solicitud de recuperación")
	}

	// 5. Enviar correo con el código de 6 dígitos
	if s.emailSvc != nil {
		go func(to, name, cod string) {
			if err := s.emailSvc.SendRecuperacionPassword(to, name, cod); err != nil {
				log.Printf("[Auth] Error enviando correo con código a %s: %v\n", to, err)
			}
		}(user.Email, user.FirstName, codigo)
	}

	return nil
}

// VerificarCodigo valida el código de 6 dígitos ingresado por el usuario y emite un token temporal de reseteo.
func (s *authService) VerificarCodigo(req *authDomain.VerificarCodigoRequest) (*authDomain.VerificarCodigoResponse, error) {
	emailBuscado := strings.TrimSpace(req.Email)
	user, err := s.userRepo.FindByEmail(emailBuscado)
	if err != nil || user == nil {
		return nil, errors.New("no se encontró ningún usuario registrado con este correo electrónico")
	}

	// Buscar el código más reciente, no usado para el usuario
	var codeRecord authDomain.PasswordResetCode
	err = s.db.Where("usuario_id = ? AND usado = false", user.ID).
		Order("id DESC").
		First(&codeRecord).Error

	if err != nil {
		return nil, errors.New("el código es inválido o ha expirado")
	}

	if time.Now().After(codeRecord.ExpiraAt) {
		return nil, errors.New("el código ha expirado")
	}

	// Validar código contra el hash
	codigoIngresado := strings.TrimSpace(req.Codigo)
	if err := bcrypt.CompareHashAndPassword([]byte(codeRecord.CodigoHash), []byte(codigoIngresado)); err != nil {
		codeRecord.IntentosFallidos++
		if codeRecord.IntentosFallidos >= 5 {
			codeRecord.Usado = true
			_ = s.db.Save(&codeRecord)
			return nil, errors.New("demasiados intentos fallidos. Solicita un nuevo código.")
		}
		_ = s.db.Save(&codeRecord)
		return nil, errors.New("código incorrecto")
	}

	// Código verificado exitosamente. Generar reset token temporal (JWT de 10 minutos)
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"purpose": "password_reset",
		"code_id": codeRecord.ID,
		"iat":     now.Unix(),
		"exp":     now.Add(10 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("error al generar token de verificación")
	}

	return &authDomain.VerificarCodigoResponse{
		ResetToken: tokenString,
		Message:    "Código verificado exitosamente.",
	}, nil
}

// ResetPassword valida el token (JWT de reseteo OTP o token de activación) y aplica la nueva contraseña al usuario.
func (s *authService) ResetPassword(req *authDomain.ResetPasswordRequest) error {
	// Intentamos primero decodificar como JWT de recuperación por código OTP
	parsedJWT, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return []byte(s.jwtSecret), nil
	})

	if err == nil && parsedJWT.Valid {
		if claims, ok := parsedJWT.Claims.(jwt.MapClaims); ok {
			if purpose, ok := claims["purpose"].(string); ok && purpose == "password_reset" {
				userIDFloat, ok := claims["user_id"].(float64)
				if !ok {
					return errors.New("token de reseteo inválido")
				}
				userID := uint(userIDFloat)

				user, err := s.userRepo.FindByID(userID)
				if err != nil || user == nil {
					return errors.New("usuario no encontrado")
				}

				nombreCompleto := user.FirstName + " " + user.LastName
				if err := validation.ValidarPassword(req.Password, user.Email, nombreCompleto); err != nil {
					return err
				}

				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
				if err != nil {
					return errors.New("falló al procesar la nueva contraseña")
				}

				tx := s.db.Begin()
				defer func() {
					if r := recover(); r != nil {
						tx.Rollback()
					}
				}()

				user.Password = string(hashedPassword)
				user.RequiereResetPassword = false
				if err := tx.Save(user).Error; err != nil {
					tx.Rollback()
					return errors.New("error al actualizar contraseña")
				}

				// Marcar el código como usado
				if codeIDFloat, ok := claims["code_id"].(float64); ok && codeIDFloat > 0 {
					_ = tx.Model(&authDomain.PasswordResetCode{}).
						Where("id = ?", uint(codeIDFloat)).
						Update("usado", true)
				} else {
					_ = tx.Model(&authDomain.PasswordResetCode{}).
						Where("usuario_id = ? AND usado = false", user.ID).
						Update("usado", true)
				}

				return tx.Commit().Error
			}
		}
	}

	// Flujo de token URL tradicional (Registro Asistido / activación de cuenta)
	var resetToken authDomain.PasswordResetToken
	if err := s.db.Where("token = ? AND usado = false", req.Token).First(&resetToken).Error; err != nil {
		return errors.New("el enlace de restablecimiento es inválido o ya ha sido utilizado")
	}

	if time.Now().After(resetToken.ExpiraAt) {
		return errors.New("el enlace de restablecimiento ha expirado")
	}

	user, err := s.userRepo.FindByID(resetToken.UsuarioID)
	if err != nil || user == nil {
		return errors.New("usuario no encontrado")
	}

	nombreCompleto := user.FirstName + " " + user.LastName
	if err := validation.ValidarPassword(req.Password, user.Email, nombreCompleto); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("falló al procesar la nueva contraseña")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user.Password = string(hashedPassword)
	user.RequiereResetPassword = false
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return errors.New("error al actualizar contraseña")
	}

	resetToken.Usado = true
	if err := tx.Save(&resetToken).Error; err != nil {
		tx.Rollback()
		return errors.New("error al actualizar el estado del token")
	}

	return tx.Commit().Error
}

// RegistroAsistido permite a un operador registrar a un cliente sin requerir ni conocer su contraseña.
// El sistema crea la cuenta con RequiereResetPassword = true y envía un correo con el token de activación.
func (s *authService) RegistroAsistido(req *authDomain.RegistroAsistidoRequest, operadorID uint) (*authDomain.RegistroAsistidoResponse, error) {
	// 1. Validar que el correo no esté registrado
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("el correo electrónico ya está registrado")
	}

	// 2. Generar contraseña temporal segura e inaccesible
	tempPassBytes := make([]byte, 24)
	if _, err := rand.Read(tempPassBytes); err != nil {
		return nil, errors.New("falló al generar credenciales temporales")
	}
	tempPass := hex.EncodeToString(tempPassBytes) + "Ab1!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("falló al procesar la contraseña del usuario")
	}

	// 3. Generar código de referido único
	codeRefer, err := generateCodeRefer()
	if err != nil {
		return nil, errors.New("falló al generar código de referencia")
	}

	// 4. Validar código de referido si se suministró
	var referente *userDomain.User
	refCodeValido := false
	if req.RefCode != nil && *req.RefCode != "" {
		ref, refErr := s.userRepo.FindByCodeRefer(*req.RefCode)
		if refErr == nil && ref != nil {
			referente = ref
			refCodeValido = true
		}
	}

	// 5. Validar empresa aliada (por código o por ID)
	var companyForLink *companyDomain.Company
	codigoEmpresaValido := false
	if req.CodigoEmpresa != nil && *req.CodigoEmpresa != "" {
		company, compErr := s.companyRepo.FindByCodigoEmpresa(*req.CodigoEmpresa)
		if compErr == nil && company != nil {
			companyForLink = company
			codigoEmpresaValido = true
		}
	} else if req.EmpresaID != nil && *req.EmpresaID > 0 {
		company, compErr := s.companyRepo.FindByID(*req.EmpresaID)
		if compErr == nil && company != nil {
			companyForLink = company
			codigoEmpresaValido = true
		}
	}

	// 6. Determinar rol
	var codeRol string
	if refCodeValido || codigoEmpresaValido {
		codeRol = "user_member"
	} else {
		codeRol = "user"
	}

	var rol roleDomain.Role
	if err := s.db.Where("code = ?", codeRol).First(&rol).Error; err != nil {
		return nil, fmt.Errorf("no se encontró el rol '%s' en el sistema", codeRol)
	}

	// 7. Crear el usuario con bandera de reseteo pendiente y auditoría del operador
	user := &userDomain.User{
		Email:                 req.Email,
		Password:              string(hashedPassword),
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		IsActive:              true,
		RoleID:                &rol.ID,
		CodeRefer:             codeRefer,
		CreateBy:              &operadorID,
		RequiereResetPassword: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 8. Asociar referido si es válido
	referidoCreated := false
	if refCodeValido && referente != nil && referente.ID != user.ID {
		convertedRef, convertErr := s.referidoSvc.ConvertirReferido(req.Email, user.ID)
		if convertErr != nil {
			nombre := req.FirstName + " " + req.LastName
			createReq := &referidoDomain.CreateReferidoRequest{
				EmailReferido:  req.Email,
				NombreReferido: &nombre,
			}
			_, createErr := s.referidoSvc.CreateReferido(createReq, referente.ID)
			if createErr == nil {
				convertedRef, _ = s.referidoSvc.ConvertirReferido(req.Email, user.ID)
			}
		}
		if convertedRef != nil {
			referidoCreated = true
		}
	}

	// 9. Asociar a empresa si es válida
	if codigoEmpresaValido && companyForLink != nil {
		_ = s.db.Exec(
			"INSERT INTO administrative.user_companies (user_id, company_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			user.ID, companyForLink.ID,
		).Error
	}

	// 10. Generar token de activación (duración 24 horas)
	token, err := generateCryptoToken()
	if err != nil {
		return nil, errors.New("falló al generar token de activación")
	}

	resetToken := &authDomain.PasswordResetToken{
		UsuarioID: user.ID,
		Token:     token,
		ExpiraAt:  time.Now().Add(24 * time.Hour),
		Usado:     false,
		Tipo:      authDomain.TipoTokenActivacion,
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		log.Printf("[Auth] Error guardando token de activación para %s: %v\n", user.Email, err)
		return nil, errors.New("error al registrar el token de activación")
	}

	// 11. Enviar correo de activación
	emailSent := false
	if s.emailSvc != nil {
		go func(to, name, tok string) {
			if err := s.emailSvc.SendActivacionCuenta(to, name, tok); err != nil {
				log.Printf("[Auth] Error enviando correo de activación a %s: %v\n", to, err)
			}
		}(user.Email, user.FirstName, token)
		emailSent = true
	}

	roleCode := rol.Code
	return &authDomain.RegistroAsistidoResponse{
		User: authDomain.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleID:    &rol.ID,
			RoleCode:  roleCode,
		},
		ReferidoCreated: referidoCreated,
		EmailSent:       emailSent,
		Message:         "Usuario registrado exitosamente. Se ha enviado un correo con el enlace para activar su cuenta.",
	}, nil
}

// generateCryptoToken genera un string hexadecimal seguro de 32 bytes aleatorios.
func generateCryptoToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateCodeRefer generates a random 8-character alphanumeric code for referrals.
func generateCodeRefer() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

// generateNumericCode genera un código numérico seguro de 6 dígitos (100000 - 999999).
func generateNumericCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	code := n.Int64() + 100000
	return fmt.Sprintf("%06d", code), nil
}


