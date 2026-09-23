package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/auth"
	authDomain "multicliente-backend/internal/features/auth/domain"
	"multicliente-backend/internal/features/benefit"
	benefitDomain "multicliente-backend/internal/features/benefit/domain"
	"multicliente-backend/internal/features/benefit_redemption"
	"multicliente-backend/internal/features/company"
	companyDomain "multicliente-backend/internal/features/company/domain"
	"multicliente-backend/internal/features/membresia"
	membresiaDomain "multicliente-backend/internal/features/membresia/domain"
	"multicliente-backend/internal/features/menu"
	menuDomain "multicliente-backend/internal/features/menu/domain"
	"multicliente-backend/internal/features/notificacion"
	notificacionDomain "multicliente-backend/internal/features/notificacion/domain"
	"multicliente-backend/internal/features/referido"
	referidoDomain "multicliente-backend/internal/features/referido/domain"
	"multicliente-backend/internal/features/rifa"
	"multicliente-backend/internal/features/role"
	roleDomain "multicliente-backend/internal/features/role/domain"
	"multicliente-backend/internal/features/upload"
	"multicliente-backend/internal/features/user"
	userDomain "multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/config"
	"multicliente-backend/internal/platform/database"
	"multicliente-backend/internal/platform/database/migrations"
	"multicliente-backend/internal/platform/database/seeds"
	"multicliente-backend/internal/platform/email"
	"multicliente-backend/internal/platform/middleware"
	"multicliente-backend/internal/platform/server"
)

func main() {
	// Load configuration from .env
	cfg := config.Load()

	// Connect to PostgreSQL
	if cfg.DatabaseURL != "" {
		log.Println("🔌 Connecting to PostgreSQL using DATABASE_URL...")
	} else {
		log.Printf("🔌 Connecting to PostgreSQL at %s:%s (database: '%s', user: '%s', sslmode: '%s')...",
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBSSLMode)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	err = migrations.Migrate(db,
		&companyDomain.Company{},
		&benefitDomain.Benefit{},
		&roleDomain.Role{},
		&roleDomain.Option{},
		&roleDomain.Permission{},
		&menuDomain.Menu{},
		&userDomain.User{},
		&referidoDomain.Referido{},
		&membresiaDomain.Membresia{},
		&membresiaDomain.PagoMembresia{},
		&membresiaDomain.Configuracion{},
		&authDomain.PasswordResetToken{},
		&authDomain.PasswordResetCode{},
		&notificacionDomain.Notificacion{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrations completed")

	// Seed database default values
	seeds.Seed(db)

	// Setup Gin router
	router := server.NewRouter()

	// API route group
	api := router.Group("/api")

	// Protected routes (JWT required)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	// Middleware to restrict endpoints to SuperAdmin role only
	superAdminRequired := middleware.RequireRole("superadmin")
	requireCompanyAccess := middleware.RequireCompanyAccess(db)

	// Email service
	emailSvc := email.NewEmailService(cfg.ResendAPIKey, cfg.ResendFromEmail, cfg.FrontendURL)

	// Notifications
	notifSvc := notificacion.RegisterRoutes(api, db, emailSvc, cfg.JWTSecret)

	// Register features
	userRepo := user.RegisterRoutes(protected, db)
	referidoSvc := referido.RegisterRoutes(protected, db)
	companyRepo := company.RegisterRoutes(protected, db, superAdminRequired)
	auth.RegisterRoutes(api, userRepo, referidoSvc, companyRepo, db, emailSvc, cfg.JWTSecret, cfg.JWTExpirationHours)
	role.RegisterRoutes(protected, db, superAdminRequired)
	menu.RegisterRoutes(protected, db, superAdminRequired)
	benefit.RegisterRoutes(protected, db, requireCompanyAccess)
	rifa.RegisterRoutes(protected, db)
	benefit_redemption.RegisterRoutes(protected, db)
	upload.RegisterRoutes(protected)

	// Register membership and payments (Wompi)
	membresiaSvc := membresia.RegisterRoutes(
		protected,
		api,
		db,
		userRepo,
		notifSvc,
		cfg.WompiPublicKey,
		cfg.WompiPrivateKey,
		cfg.WompiEventsSecret,
		cfg.WompiIntegritySecret,
		cfg.WompiSandbox,
	)

	// Start background cron jobs (trial expiration & auto-renewals)
	membresia.StartDailyCronJobs(db, membresiaSvc)

	// Health check (public)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
