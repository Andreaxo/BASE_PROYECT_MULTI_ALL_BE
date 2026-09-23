package database

import (
	"fmt"
	"strings"

	"multicliente-backend/internal/platform/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL using GORM.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dsn string
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		escapedPassword := strings.ReplaceAll(cfg.DBPassword, "'", "\\'")
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password='%s' dbname=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, escapedPassword, cfg.DBName, cfg.DBSSLMode,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
