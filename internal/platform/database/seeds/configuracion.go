package seeds

import (
	"gorm.io/gorm"
)

// SeedConfiguracion populates default configuration entries if they don't exist.
func SeedConfiguracion(db *gorm.DB) error {
	configs := []struct {
		Clave       string
		Valor       string
		Descripcion string
	}{
		{
			Clave:       "meses_gratis_negocio",
			Valor:       "12",
			Descripcion: "Meses de prueba gratuita para negocios aliados",
		},
		{
			Clave:       "precio_membresia_mensual",
			Valor:       "25000",
			Descripcion: "Precio mensual de membresía para afiliados en COP",
		},
	}

	for _, c := range configs {
		query := `
			INSERT INTO administrative.configuracion (clave, valor, descripcion)
			VALUES (?, ?, ?)
			ON CONFLICT (clave) DO NOTHING;
		`
		if err := db.Exec(query, c.Clave, c.Valor, c.Descripcion).Error; err != nil {
			return err
		}
	}

	return nil
}
