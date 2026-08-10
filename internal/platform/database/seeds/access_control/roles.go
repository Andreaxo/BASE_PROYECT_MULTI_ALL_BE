package access_control

import (
	"log"
	"gorm.io/gorm"
	roleDomain "multicliente-backend/internal/features/role/domain"
)

// SeedRoles seeds system roles into database.
func SeedRoles(db *gorm.DB) error {
	roles := []roleDomain.Role{
		{ID: 1, Name: "Super Administrador", Code: "superadmin", Description: "Super Administrador del sistema con acceso total.", SessionDays: 30, SessionHours: 0, SessionMinutes: 0, IsActive: true},
		{ID: 2, Name: "Administrador", Code: "admin", Description: "Administrador de la empresa con acceso a sucursales.", SessionDays: 7, SessionHours: 0, SessionMinutes: 0, IsActive: true},
		{ID: 3, Name: "Usuario", Code: "user", Description: "Usuario de la empresa con permisos estándar de inventario.", SessionDays: 1, SessionHours: 0, SessionMinutes: 0, IsActive: true},
		{ID: 7, Name: "Negocio", Code: "business_validator", Description: "Usuario validador de redenciones de beneficios de una empresa aliada", SessionDays: 1, SessionHours: 0, SessionMinutes: 0, IsActive: true},
	}
	for _, r := range roles {
		var existing roleDomain.Role
		if err := db.Where("id = ?", r.ID).First(&existing).Error; err != nil {
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		} else {
			existing.Name = r.Name
			existing.Code = r.Code
			existing.Description = r.Description
			existing.SessionDays = r.SessionDays
			existing.SessionHours = r.SessionHours
			existing.SessionMinutes = r.SessionMinutes
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
		}
	}
	log.Println("✅ Roles seeded")
	return nil
}
