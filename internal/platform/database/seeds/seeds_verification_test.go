package seeds_test

import (
	"testing"

	"multicliente-backend/internal/features/menu/domain"
	roleDomain "multicliente-backend/internal/features/role/domain"
	"multicliente-backend/internal/platform/config"
	"multicliente-backend/internal/platform/database"
	"multicliente-backend/internal/platform/database/seeds/access_control"
)

func TestSeedsVerification(t *testing.T) {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	t.Log("1. Running SeedRoles...")
	if err := access_control.SeedRoles(db); err != nil {
		t.Fatalf("SeedRoles failed: %v", err)
	}

	t.Log("2. Running SeedMenus...")
	if err := access_control.SeedMenus(db); err != nil {
		t.Fatalf("SeedMenus failed: %v", err)
	}

	t.Log("3. Running SeedPermissions...")
	if err := access_control.SeedPermissions(db); err != nil {
		t.Fatalf("SeedPermissions failed: %v", err)
	}

	// Assertion 1: user_member exists and is active
	t.Run("Verify user_member role seeded", func(t *testing.T) {
		var role roleDomain.Role
		err := db.Where("code = ?", "user_member").First(&role).Error
		if err != nil {
			t.Fatalf("CRITICAL: role 'user_member' NOT found in DB: %v", err)
		}
		if role.ID != 5 {
			t.Errorf("Expected role ID 5 for user_member, got %d", role.ID)
		}
		if !role.IsActive {
			t.Errorf("Expected user_member to be active, got false")
		}
		t.Logf("✅ user_member found: ID=%d, Name='%s', Code='%s', Active=%v", role.ID, role.Name, role.Code, role.IsActive)
	})

	// Assertion 2: Role lookup simulation as done in auth registration
	t.Run("Verify auth registration role lookup for affiliated users", func(t *testing.T) {
		codeRol := "user_member"
		var rol roleDomain.Role
		err := db.Where("code = ?", codeRol).First(&rol).Error
		if err != nil {
			t.Fatalf("Registration lookup failed for code '%s': %v", codeRol, err)
		}
		t.Logf("✅ Auth registration query 'WHERE code = user_member' succeeded. Role ID: %d", rol.ID)
	})

	// Assertion 3: Orphan role 'empresa' (ID 4) does NOT exist
	t.Run("Verify orphan role 'empresa' is deleted/absent", func(t *testing.T) {
		var count int64
		db.Model(&roleDomain.Role{}).Where("id = 4 OR code = 'empresa'").Count(&count)
		if count > 0 {
			t.Fatalf("Expected role 'empresa' to be eliminated, but found %d records", count)
		}
		t.Log("✅ Orphan role 'empresa' (ID 4) is confirmed absent from DB")
	})

	// Assertion 4: Permissions for user_member (RoleID 5)
	t.Run("Verify permissions for user_member", func(t *testing.T) {
		var permCount int64
		db.Model(&roleDomain.Permission{}).Where("role_id = ?", 5).Count(&permCount)
		if permCount == 0 {
			t.Fatalf("CRITICAL: user_member has 0 permissions seeded!")
		}
		t.Logf("✅ user_member has %d permission records in DB", permCount)

		// Check menu access: 11 (Referidos), 12 (Beneficios), 13 (Rifas)
		for _, menuID := range []uint{11, 12, 13} {
			var menuPerms int64
			db.Model(&roleDomain.Permission{}).Where("role_id = 5 AND menu_id = ?", menuID).Count(&menuPerms)
			if menuPerms == 0 {
				t.Errorf("Missing permissions for role 5 on menu %d", menuID)
			} else {
				t.Logf("✅ user_member has %d permissions for Menu %d", menuPerms, menuID)
			}
		}
	})

	t.Log("4. Running SeedUsers...")
	if err := access_control.SeedUsers(db); err != nil {
		t.Fatalf("SeedUsers failed: %v", err)
	}

	// Assertion 5: Menus 5, 6, 7 are deactivated (IsActive = false)
	t.Run("Verify inventory menus 5, 6, 7 are deactivated", func(t *testing.T) {
		for _, menuID := range []uint{5, 6, 7} {
			var m domain.Menu
			if err := db.Where("id = ?", menuID).First(&m).Error; err == nil {
				if m.IsActive {
					t.Errorf("Expected Menu %d ('%s') to have IsActive=false, got true", menuID, m.Label)
				} else {
					t.Logf("✅ Menu %d ('%s') is correctly deactivated (IsActive=false)", menuID, m.Label)
				}
			}
		}
	})

	// Assertion 6: Operador role, user and permissions
	t.Run("Verify operador role, user, and permissions", func(t *testing.T) {
		var role roleDomain.Role
		if err := db.Where("code = ?", "operador").First(&role).Error; err != nil {
			t.Fatalf("CRITICAL: role 'operador' NOT found in DB: %v", err)
		}
		t.Logf("✅ operador role found: ID=%d, Name='%s'", role.ID, role.Name)

		var userCount int64
		db.Table("administrative.users").Where("email = ? AND role_id = 8", "operador@conexiate.com").Count(&userCount)
		if userCount == 0 {
			t.Fatalf("CRITICAL: test operador user 'operador@conexiate.com' not found in DB")
		}
		t.Logf("✅ test operador user found in DB")

		var permCount int64
		db.Model(&roleDomain.Permission{}).Where("role_id = 8").Count(&permCount)
		t.Logf("✅ operador has %d permissions configured", permCount)
	})
}
