package application

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	authDomain "multicliente-backend/internal/features/auth/domain"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	userDomain "multicliente-backend/internal/features/user/domain"
)

type authService struct {
	userRepo       userDomain.UserRepository
	referidoSvc    referidoDomain.ReferidoService
	jwtSecret      string
	jwtExpHours    int
}

// NewAuthService creates a new AuthService.
// It depends on the user repository (cross-feature dependency via domain port)
// and the referido service for handling referral logic during registration.
func NewAuthService(
	userRepo userDomain.UserRepository,
	referidoSvc referidoDomain.ReferidoService,
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
// If ref_code is provided, it links the referral record after user creation.
// All operations happen in the service layer; DB transaction is handled by the repository.
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

	// 4. Assign default role (user = ID 3)
	defaultRoleID := uint(3)

	// 5. Create the user
	user := &userDomain.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		RoleID:    &defaultRoleID,
		CodeRefer: codeRefer,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 6. Handle referral logic if ref_code was provided
	referidoCreated := false
	if req.RefCode != nil && *req.RefCode != "" {
		referente, refErr := s.userRepo.FindByCodeRefer(*req.RefCode)
		if refErr == nil && referente != nil {
			// Prevent self-referral
			if referente.ID != user.ID {
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
				// Auto-affiliate: transition from 'registrado' → 'afiliado' automatically
				if convertedRef != nil {
					_, _ = s.referidoSvc.AfiliarReferido(convertedRef.ID)
					referidoCreated = true
				}
			}
		}
		// If ref_code is invalid, continue registration without breaking
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
