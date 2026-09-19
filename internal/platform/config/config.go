package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values loaded from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret          string
	JWTExpirationHours string

	ServerPort string

	// Wompi payment gateway
	WompiPublicKey    string
	WompiPrivateKey   string
	WompiEventsSecret string
	WompiSandbox      bool
}

// Load reads the .env file and returns a Config struct with all values.
func Load() *Config {
	// Load .env file (ignore error if not found)
	godotenv.Load()

	sandbox := getEnv("WOMPI_SANDBOX", "true")

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "localhost"),
		DBUser:     getEnv("DB_USER", "localhost"),
		DBPassword: getEnv("DB_PASSWORD", "localhost"),
		DBName:     getEnv("DB_NAME", "localhost"),
		DBSSLMode:  getEnv("DB_SSLMODE", "localhost"),

		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-key-change-me-in-production"),
		JWTExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),

		ServerPort: getEnv("SERVER_PORT", "8080"),

		WompiPublicKey:    getEnv("WOMPI_PUBLIC_KEY", ""),
		WompiPrivateKey:   getEnv("WOMPI_PRIVATE_KEY", ""),
		WompiEventsSecret: getEnv("WOMPI_EVENTS_SECRET", ""),
		WompiSandbox:      sandbox == "true" || sandbox == "1",
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
