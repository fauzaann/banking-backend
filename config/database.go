package config

import "fmt"

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	TimeZone     string
	MaxOpenConns int
	MaxIdleConns int
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnv("DB_PORT", "5432"),
		User:         getEnv("DB_USER", "postgres"),
		Password:     getEnv("DB_PASSWORD", "postgres"),
		Name:         getEnv("DB_NAME", "banking_db"),
		SSLMode:      getEnv("DB_SSLMODE", "disable"),
		TimeZone:     getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
	}
}

// DSN membentuk connection string PostgreSQL.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode, d.TimeZone,
	)
}
