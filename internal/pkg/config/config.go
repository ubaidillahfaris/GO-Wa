package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	JWT      JWTConfig
	WhatsApp WhatsAppConfig
	CORS     CORSConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	URI      string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	ExpiresMin int
}

// WhatsAppConfig holds WhatsApp configuration
type WhatsAppConfig struct {
	StoresDir      string
	UploadsDir     string
	MaxConcurrency int
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	MaxAge         int
}

var cfg *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "3000"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ShutdownTimeout: 10 * time.Second,
		},
		MongoDB: MongoDBConfig{
			User:     getEnv("MONGO_USER", ""),
			Password: getEnv("MONGO_PASS", ""),
			Host:     getEnv("MONGO_HOST", ""),
			Port:     getEnv("MONGO_PORT", "27017"),
			Database: getEnv("MONGO_DB", "qr_db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "mysecretkey"),
			ExpiresMin: getEnvAsInt("JWT_EXPIRES_MIN", 60),
		},
		WhatsApp: WhatsAppConfig{
			StoresDir:      getEnv("WHATSAPP_STORES_DIR", "./stores"),
			UploadsDir:     getEnv("WHATSAPP_UPLOADS_DIR", "./uploads/whatsapp"),
			MaxConcurrency: getEnvAsInt("WHATSAPP_MAX_CONCURRENCY", 10),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{
				getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			MaxAge:         getEnvAsInt("CORS_MAX_AGE", 43200),
		},
	}

	// Build MongoDB URI
	config.MongoDB.URI = buildMongoURI(config.MongoDB)

	// Validate configuration
	if err := validate(config); err != nil {
		return nil, err
	}

	cfg = config
	return config, nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		panic("Configuration not loaded. Call Load() first.")
	}
	return cfg
}

// buildMongoURI constructs MongoDB connection URI
func buildMongoURI(mongo MongoDBConfig) string {
	if mongo.Host == "" {
		// Auto-construct from PORT if not set (for docker-compose)
		serverPort := getEnv("PORT", "3000")
		mongoPort, _ := strconv.Atoi(serverPort)
		mongoPort += 10000
		mongo.Host = fmt.Sprintf("mongo:%d", mongoPort)
	}

	if mongo.User != "" && mongo.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s",
			mongo.User,
			mongo.Password,
			mongo.Host,
		)
	}

	return fmt.Sprintf("mongodb://%s", mongo.Host)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if config.MongoDB.Database == "" {
		return fmt.Errorf("MONGO_DB is required")
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// IsDevelopment checks if running in development mode
func IsDevelopment() bool {
	return Get().Server.Environment == "development"
}

// IsProduction checks if running in production mode
func IsProduction() bool {
	return Get().Server.Environment == "production"
}
