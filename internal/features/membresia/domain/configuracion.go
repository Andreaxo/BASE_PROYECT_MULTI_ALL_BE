package domain

import "gorm.io/gorm"

// Configuracion maps to the administrative.configuracion table.
type Configuracion struct {
	Clave       string `gorm:"primaryKey;type:varchar(50)" json:"clave"`
	Valor       string `gorm:"type:varchar(200);not null" json:"valor"`
	Descripcion string `gorm:"type:varchar(300)" json:"descripcion"`
}

func (Configuracion) TableName() string {
	return "administrative.configuracion"
}

// GetConfigValue reads a configuration value by key from the database.
// Returns the value as string, or defaultVal if not found.
func GetConfigValue(db *gorm.DB, clave string, defaultVal string) string {
	var cfg Configuracion
	if err := db.Where("clave = ?", clave).First(&cfg).Error; err != nil {
		return defaultVal
	}
	return cfg.Valor
}
