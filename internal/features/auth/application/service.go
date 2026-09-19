package application

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authDomain "multicliente-backend/internal/features/auth/domain"
	companyDomain "multicliente-backend/internal/features/company/domain"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
)

type authService struct {
	userRepo    userDomain.UserRepository
	referidoSvc referidoDomain.ReferidoService
	companyRepo companyDomain.CompanyRepository
	db          *gorm.DB
	jwtSecret   string
	jwtExpHours int
}

// NewAuthService creates a new AuthService.
// It depends on the user repository, referido service, company repository, and DB
// for handling referral logic, company code linking, and dynamic role assignment during registration.
func NewAuthService(
	userRepo userDomain.UserRepository,
	referidoSvc referidoDomain.ReferidoService,
	companyRepo companyDomain.CompanyRepository,
	db *gorm.DB,
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

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
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

	// 2. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("falló al procesar la contraseña")
	}

	// 3. Generate unique code_refer for the new user
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
