package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/joho/godotenv"
)

const defaultDBParams = "parseTime=true&charset=utf8mb4&timeout=10s&readTimeout=60s&writeTimeout=30s"

type AppConfig struct {
	Server    ServerConfig
	Database  DBConfig
	Firestore FirestoreConfig
	Timeouts  TimeoutsConfig
	Env       string
	External  ExternalConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type TimeoutsConfig struct {
	ItemUpdate time.Duration
}

type DBConfig struct {
	User            string
	Password        string
	DBName          string
	Host            string
	Port            string
	UnixSocket      string
	Params          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type FirestoreConfig struct {
	ProjectID       string
	DatabaseID      string
	CredentialsFile string
	ItemsCollection string
	SecretName      string
}

type ExternalConfig struct {
	ServiceURL string
}

// LoadConfig carga la configuración de la aplicación
func LoadConfig() (*AppConfig, error) {
	godotenv.Load()

	// Cloud Run usa la variable PORT, pero localmente usamos SERVER_PORT
	serverPort := getEnv("PORT", "")
	if serverPort == "" {
		serverPort = getEnv("SERVER_PORT", "8080")
	}

	cfg := &AppConfig{
		Env: getEnv("APP_ENV", "dev"),
		Server: ServerConfig{
			Port:         serverPort,
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 120*time.Second),
		},
		Timeouts: TimeoutsConfig{
			ItemUpdate: getEnvDuration("ITEM_UPDATE_TIMEOUT", 60*time.Second),
		},
		Database: DBConfig{
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", ""),
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			UnixSocket:      getEnv("DB_UNIX_SOCKET", ""),
			Params:          getEnv("DB_PARAMS", defaultDBParams),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Firestore: FirestoreConfig{
			ProjectID:       getEnv("GCP_PROJECT_ID", ""),
			DatabaseID:      getEnv("FIRESTORE_DATABASE_ID", "logs"),
			CredentialsFile: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			ItemsCollection: "items-data-collection",
			SecretName:      getEnv("FIREBASE_SECRET_NAME", ""),
		},
		External: ExternalConfig{
			ServiceURL: getEnv("EXTERNAL_SERVICE_URL", ""),
		},
	}

	return cfg, nil
}

// DSN construye el Data Source Name para MySQL
// Soporta conexión por TCP (local) y Unix Socket (Cloud SQL)
func (db *DBConfig) DSN() string {
	if db.UnixSocket != "" {
		// Conexión via Unix Socket (Cloud SQL en Cloud Run)
		return fmt.Sprintf("%s:%s@unix(%s)/%s?%s",
			db.User,
			db.Password,
			db.UnixSocket,
			db.DBName,
			db.Params,
		)
	}
	// Conexión via TCP (local/desarrollo)
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.DBName,
		db.Params,
	)
}

// SafeSummary devuelve un resumen de la configuración sin exponer la contraseña
func (db *DBConfig) SafeSummary() string {
	if db.UnixSocket != "" {
		return fmt.Sprintf("unix(%s)/%s MaxOpen=%d MaxIdle=%d",
			db.UnixSocket, db.DBName, db.MaxOpenConns, db.MaxIdleConns)
	}
	return fmt.Sprintf("tcp(%s:%s)/%s MaxOpen=%d MaxIdle=%d",
		db.Host, db.Port, db.DBName, db.MaxOpenConns, db.MaxIdleConns)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetSecretValue obtiene un valor desde Secret Manager
func GetSecretValue(ctx context.Context, projectID, secretName string) ([]byte, error) {
	// Limpiar GOOGLE_APPLICATION_CREDENTIALS temporalmente para forzar el uso
	// del service account de Cloud Run
	oldCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if oldCreds != "" {
		os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		defer os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", oldCreds)
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer client.Close()

	// Construir el nombre del recurso
	resourceName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName)

	// Acceder al secret
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %w", err)
	}

	return result.Payload.Data, nil
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
