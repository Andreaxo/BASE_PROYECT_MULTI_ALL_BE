package access_control

import (
	"log"

	"gorm.io/gorm"

	menuDomain "multicliente-backend/internal/features/menu/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
)

// SeedPermissions seeds default CRUD permissions for superadmin.
func SeedPermissions(db *gorm.DB) error {
	permissions := []roleDomain.Permission{
		{RoleID: 1, MenuID: 1, OptionID: 1},
		{RoleID: 1, MenuID: 1, OptionID: 2},
		{RoleID: 1, MenuID: 1, OptionID: 3},
		{RoleID: 1, MenuID: 1, OptionID: 4},
		{RoleID: 1, MenuID: 2, OptionID: 1},
		{RoleID: 1, MenuID: 2, OptionID: 2},
		{RoleID: 1, MenuID: 2, OptionID: 3},
		{RoleID: 1, MenuID: 2, OptionID: 4},
		{RoleID: 1, MenuID: 3, OptionID: 1},
		{RoleID: 1, MenuID: 3, OptionID: 2},
		{RoleID: 1, MenuID: 3, OptionID: 3},
		{RoleID: 1, MenuID: 3, OptionID: 4},
		{RoleID: 1, MenuID: 4, OptionID: 1},
		{RoleID: 1, MenuID: 4, OptionID: 2},
		{RoleID: 1, MenuID: 4, OptionID: 3},
		{RoleID: 1, MenuID: 4, OptionID: 4},
		{RoleID: 1, MenuID: 5, OptionID: 1},
		{RoleID: 1, MenuID: 5, OptionID: 2},
		{RoleID: 1, MenuID: 5, OptionID: 3},
		{RoleID: 1, MenuID: 5, OptionID: 4},
		{RoleID: 1, MenuID: 6, OptionID: 1},
		{RoleID: 1, MenuID: 6, OptionID: 2},
		{RoleID: 1, MenuID: 6, OptionID: 3},
		{RoleID: 1, MenuID: 6, OptionID: 4},
		{RoleID: 1, MenuID: 7, OptionID: 1},
		{RoleID: 1, MenuID: 7, OptionID: 2},
		{RoleID: 1, MenuID: 7, OptionID: 3},
		{RoleID: 1, MenuID: 7, OptionID: 4},
		// Permissions for Access Control Folder itself (MenuID 8)
		{RoleID: 1, MenuID: 8, OptionID: 1},
		{RoleID: 1, MenuID: 8, OptionID: 2},
		{RoleID: 1, MenuID: 8, OptionID: 3},
		{RoleID: 1, MenuID: 8, OptionID: 4},
		// Permissions for Statistics (MenuID 9)
		{RoleID: 1, MenuID: 9, OptionID: 1},
		{RoleID: 1, MenuID: 9, OptionID: 2},
		{RoleID: 1, MenuID: 9, OptionID: 3},
		{RoleID: 1, MenuID: 9, OptionID: 4},
		// Permissions for Dashboard (MenuID 10)
		{RoleID: 1, MenuID: 10, OptionID: 1},
		{RoleID: 1, MenuID: 10, OptionID: 2},
		{RoleID: 1, MenuID: 10, OptionID: 3},
		{RoleID: 1, MenuID: 10, OptionID: 4},
		// Permissions for Referidos (MenuID 11)
		{RoleID: 1, MenuID: 11, OptionID: 1},
		{RoleID: 1, MenuID: 11, OptionID: 2},
		{RoleID: 1, MenuID: 11, OptionID: 3},
		{RoleID: 1, MenuID: 11, OptionID: 4},
		{RoleID: 2, MenuID: 11, OptionID: 1},
		{RoleID: 2, MenuID: 11, OptionID: 2},
		{RoleID: 2, MenuID: 11, OptionID: 3},
		{RoleID: 2, MenuID: 11, OptionID: 4},
		{RoleID: 3, MenuID: 11, OptionID: 1},
		{RoleID: 3, MenuID: 11, OptionID: 2},
		{RoleID: 5, MenuID: 11, OptionID: 1},
		{RoleID: 5, MenuID: 11, OptionID: 2},
		// Permissions for Beneficios (MenuID 12)
		{RoleID: 1, MenuID: 12, OptionID: 1},
		{RoleID: 1, MenuID: 12, OptionID: 2},
		{RoleID: 1, MenuID: 12, OptionID: 3},
		{RoleID: 1, MenuID: 12, OptionID: 4},
		{RoleID: 2, MenuID: 12, OptionID: 1},
		{RoleID: 2, MenuID: 12, OptionID: 2},
		{RoleID: 2, MenuID: 12, OptionID: 3},
		{RoleID: 2, MenuID: 12, OptionID: 4},
		{RoleID: 3, MenuID: 12, OptionID: 1},
		{RoleID: 3, MenuID: 12, OptionID: 2},
		{RoleID: 5, MenuID: 12, OptionID: 1},
		{RoleID: 5, MenuID: 12, OptionID: 2},
		// Permissions for Rifas (MenuID 13)
		{RoleID: 1, MenuID: 13, OptionID: 1},
		{RoleID: 1, MenuID: 13, OptionID: 2},
		{RoleID: 1, MenuID: 13, OptionID: 3},
		{RoleID: 1, MenuID: 13, OptionID: 4},
		{RoleID: 2, MenuID: 13, OptionID: 1},
		{RoleID: 2, MenuID: 13, OptionID: 2},
		{RoleID: 2, MenuID: 13, OptionID: 3},
		{RoleID: 2, MenuID: 13, OptionID: 4},
		{RoleID: 3, MenuID: 13, OptionID: 1},
		{RoleID: 3, MenuID: 13, OptionID: 2},
		{RoleID: 5, MenuID: 13, OptionID: 1},
		{RoleID: 5, MenuID: 13, OptionID: 2},
		// Permissions for Operador (RoleID 8)
		// Access Control folder
		{RoleID: 8, MenuID: 8, OptionID: 1},
		// Usuarios
		{RoleID: 8, MenuID: 1, OptionID: 1},
		{RoleID: 8, MenuID: 1, OptionID: 2},
		{RoleID: 8, MenuID: 1, OptionID: 3},
		// Negocios Aliados
		{RoleID: 8, MenuID: 2, OptionID: 1},
		{RoleID: 8, MenuID: 2, OptionID: 2},
		{RoleID: 8, MenuID: 2, OptionID: 3},
		// Estadísticas
		{RoleID: 8, MenuID: 9, OptionID: 1},
		// Dashboard
		{RoleID: 8, MenuID: 10, OptionID: 1},
		// Referidos
		{RoleID: 8, MenuID: 11, OptionID: 1},
		{RoleID: 8, MenuID: 11, OptionID: 2},
		{RoleID: 8, MenuID: 11, OptionID: 3},
		// Beneficios
		{RoleID: 8, MenuID: 12, OptionID: 1},
		{RoleID: 8, MenuID: 12, OptionID: 2},
		{RoleID: 8, MenuID: 12, OptionID: 3},
		// Rifas (VIEW and EDIT only, NO CREATE)
		{RoleID: 8, MenuID: 13, OptionID: 1},
		{RoleID: 8, MenuID: 13, OptionID: 2},
	}
	for _, p := range permissions {
		var existingMenu menuDomain.Menu
		if err := db.Where("id = ?", p.MenuID).First(&existingMenu).Error; err == nil {
			if err := db.FirstOrCreate(&p, roleDomain.Permission{RoleID: p.RoleID, MenuID: p.MenuID, OptionID: p.OptionID}).Error; err != nil {
				return err
			}
		}
	}
	log.Println("✅ Permissions seeded")
	return nil
}
