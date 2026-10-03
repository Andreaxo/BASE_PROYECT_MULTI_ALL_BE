package membresia

import (
	"log"
	"time"

	"gorm.io/gorm"

	"multicliente-backend/internal/features/membresia/domain"
)

// ExecuteDailyJobs runs the periodic subscription expiration and membership renewal routines.
func ExecuteDailyJobs(db *gorm.DB, membSvc domain.MembresiaService) {
	log.Println("⏰ Running daily subscription & membership jobs...")

	// 1. Expire company trial subscriptions that have ended
	result := db.Exec(`
		UPDATE administrative.companies
		SET suscripcion_estado = 'vencida'
		WHERE suscripcion_estado = 'prueba'
		  AND fecha_fin_prueba IS NOT NULL
		  AND fecha_fin_prueba < NOW()
	`)
	if result.Error != nil {
		log.Printf("⚠️ Error expiring trial company subscriptions: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("ℹ️ Expired %d company trial subscription(s)", result.RowsAffected)
	}

	// 2. Automatically disable benefits of companies whose subscription has ended or is unpaid
	resBenefits := db.Exec(`
		UPDATE administrative.benefit
		SET is_active = false, estado = 'deshabilitado'
		WHERE company_benefit IN (
			SELECT id FROM administrative.companies
			WHERE suscripcion_estado IN ('vencida', 'suspendida')
			   OR (suscripcion_estado = 'prueba' AND fecha_fin_prueba IS NOT NULL AND fecha_fin_prueba < NOW())
		)
		  AND estado != 'deshabilitado'
	`)
	if resBenefits.Error != nil {
		log.Printf("⚠️ Error disabling benefits for unpaid companies: %v", resBenefits.Error)
	} else if resBenefits.RowsAffected > 0 {
		log.Printf("ℹ️ Automatically disabled %d benefit(s) belonging to unpaid companies", resBenefits.RowsAffected)
	}

	// 2. Process automatic renewals for expired active memberships
	if err := membSvc.RenovarVencidas(); err != nil {
		log.Printf("⚠️ Error renewing expired memberships: %v", err)
	}

	log.Println("✅ Daily subscription & membership jobs finished")
}

// StartDailyCronJobs starts a background goroutine that runs daily jobs on startup and every 24 hours.
func StartDailyCronJobs(db *gorm.DB, membSvc domain.MembresiaService) {
	go func() {
		// Run once on server startup
		ExecuteDailyJobs(db, membSvc)

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ExecuteDailyJobs(db, membSvc)
		}
	}()
}
