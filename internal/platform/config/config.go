package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values loaded from environment variables.
type Config struct {
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string

	JWTSecret          string
	JWTExpirationHours string

	ServerPort string

	// Wompi payment gateway
	WompiPublicKey       string
	WompiPrivateKey      string
	WompiEventsSecret    string
	WompiIntegritySecret string
	WompiSandbox         bool

	// Resend email service
	ResendAPIKey    string
	ResendFromEmail string

	// Frontend URL for email links
	FrontendURL string
}

// Load reads the .env file and returns a Config struct with all values.
func Load() *Config {
	// Load .env file (ignore error if not found)
	godotenv.Load()

	sandbox := getEnv("WOMPI_SANDBOX", "true")

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "conexiate_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-key-change-me-in-production"),
		JWTExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),

		ServerPort: getEnv("SERVER_PORT", "8080"),

		WompiPublicKey:       getEnv("WOMPI_PUBLIC_KEY", ""),
		WompiPrivateKey:      getEnv("WOMPI_PRIVATE_KEY", ""),
		WompiEventsSecret:    getEnv("WOMPI_EVENTS_SECRET", ""),
		WompiIntegritySecret: getEnv("WOMPI_INTEGRITY_SECRET", ""),
		WompiSandbox:         sandbox == "true" || sandbox == "1",

		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
