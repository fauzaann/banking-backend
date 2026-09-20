package config
import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menampung seluruh konfigurasi aplikasi yang dibaca dari environment.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

// Load membaca file .env (jika ada) lalu memetakan environment ke struct Config.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: loadDatabaseConfig(),
		JWT:      loadJWTConfig(),
	}

	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "change-this-secret" && cfg.IsProduction() {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi dengan nilai yang aman")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("DB_NAME wajib diisi")
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.App.Env == "production" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
