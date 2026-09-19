package seeds

import (
	"log"
	"gorm.io/gorm"
	"multicliente-backend/internal/platform/database/seeds/access_control"
)

// Seed populates default options, roles, menus, permissions, companies, and the admin user.
func Seed(db *gorm.DB) {
	if err := access_control.SeedOptions(db); err != nil {
		log.Printf("❌ Failed to seed options: %v", err)
		return
	}
	if err := access_control.SeedRoles(db); err != nil {
		log.Printf("❌ Failed to seed roles: %v", err)
		return
	}
	if err := access_control.SeedMenus(db); err != nil {
		log.Printf("❌ Failed to seed menus: %v", err)
		return
	}
	if err := access_control.SeedPermissions(db); err != nil {
		log.Printf("❌ Failed to seed permissions: %v", err)
		return
	}
	if err := SeedCompanies(db); err != nil {
		log.Printf("❌ Failed to seed companies: %v", err)
		return
	}
	if err := access_control.SeedUsers(db); err != nil {
		log.Printf("❌ Failed to seed users: %v", err)
		return
	}
	if err := SeedBenefits(db); err != nil {
		log.Printf("❌ Failed to seed benefits: %v", err)
		return
	}
	if err := SeedConfiguracion(db); err != nil {
		log.Printf("❌ Failed to seed configuracion: %v", err)
		return
	}

	// Reset sequences in Postgres for manually inserted IDs
	db.Exec("SELECT setval(pg_get_serial_sequence('administrative.roles', 'id'), COALESCE((SELECT MAX(id) FROM administrative.roles), 1))")
	db.Exec("SELECT setval(pg_get_serial_sequence('administrative.menus', 'id'), COALESCE((SELECT MAX(id) FROM administrative.menus), 1))")
	db.Exec("SELECT setval(pg_get_serial_sequence('administrative.companies', 'id'), COALESCE((SELECT MAX(id) FROM administrative.companies), 1))")
	db.Exec("SELECT setval(pg_get_serial_sequence('administrative.benefit', 'id'), COALESCE((SELECT MAX(id) FROM administrative.benefit), 1))")
	log.Println("✅ PK Sequences synced successfully")
}
