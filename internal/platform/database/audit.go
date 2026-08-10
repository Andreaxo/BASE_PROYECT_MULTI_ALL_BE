package database

import (
	"gorm.io/gorm"
)

type AuditUser struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	EmpresaID uint   `json:"empresa_id"`
}

// GetUserNamesMap queries the database for user names and returns a map of ID -> FullName.
func GetUserNamesMap(db *gorm.DB, userIDs []uint) (map[uint]string, error) {
	if len(userIDs) == 0 {
		return map[uint]string{}, nil
	}

	// De-duplicate user IDs
	uniqueIDsMap := make(map[uint]bool)
	var uniqueIDs []uint
	for _, id := range userIDs {
		if !uniqueIDsMap[id] {
			uniqueIDsMap[id] = true
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	var users []AuditUser
	if err := db.Table("administrative.users").
		Select("id, first_name, last_name, empresa_id").
		Where("id IN ?", uniqueIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}

	namesMap := make(map[uint]string)
	for _, u := range users {
		name := u.FirstName + " " + u.LastName
		namesMap[u.ID] = name
	}
	return namesMap, nil
}
