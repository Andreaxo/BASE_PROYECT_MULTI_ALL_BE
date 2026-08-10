package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/article"
	articleDomain "multicliente-backend/internal/features/article/domain"
	"multicliente-backend/internal/features/auth"
	"multicliente-backend/internal/features/benefit"
	benefitDomain "multicliente-backend/internal/features/benefit/domain"
	"multicliente-backend/internal/features/benefit_redemption"
	"multicliente-backend/internal/features/category"
	categoryDomain "multicliente-backend/internal/features/category/domain"
	"multicliente-backend/internal/features/company"
	companyDomain "multicliente-backend/internal/features/company/domain"
	"multicliente-backend/internal/features/menu"
	menuDomain "multicliente-backend/internal/features/menu/domain"
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
	"multicliente-backend/internal/platform/middleware"
	"multicliente-backend/internal/platform/server"
)

func main() {
	// Load configuration from .env
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// // Run auto-migrations in order of dependency
	// // Note: Drop legacy UUID columns if they exist in administrative.users, as PostgreSQL cannot alter UUID to bigint directly
	// _ = db.Exec("ALTER TABLE administrative.users DROP COLUMN IF EXISTS create_by")
	// _ = db.Exec("ALTER TABLE administrative.users DROP COLUMN IF EXISTS update_by")
	// // Safety: drop leftover constraints that GORM may try to manage but don't exist in DB yet
	// _ = db.Exec(`ALTER TABLE "administrative"."companies" DROP CONSTRAINT IF EXISTS "uni_companies_nit"`)

	err = migrations.Migrate(db,
		&companyDomain.Company{},
		&benefitDomain.Benefit{},
		&categoryDomain.Category{},
		&articleDomain.Article{},
		&roleDomain.Role{},
		&roleDomain.Option{},
		&roleDomain.Permission{},
		&menuDomain.Menu{},
		&userDomain.User{},
		&referidoDomain.Referido{},
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

	// Register features
	userRepo := user.RegisterRoutes(protected, db)
	referidoSvc := referido.RegisterRoutes(protected, db)
	auth.RegisterRoutes(api, userRepo, referidoSvc, cfg.JWTSecret, cfg.JWTExpirationHours)
	company.RegisterRoutes(protected, db, superAdminRequired)
	role.RegisterRoutes(protected, db, superAdminRequired)
	menu.RegisterRoutes(protected, db, superAdminRequired)
	benefit.RegisterRoutes(protected, db, requireCompanyAccess)
	category.RegisterRoutes(protected, db, requireCompanyAccess)
	article.RegisterRoutes(protected, db, requireCompanyAccess)
	rifa.RegisterRoutes(protected, db)
	benefit_redemption.RegisterRoutes(protected, db)
	upload.RegisterRoutes(protected)

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
