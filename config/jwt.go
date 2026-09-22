package config

import "time"

type JWTConfig struct {
	Secret             string
	ExpireHours        int
	RefreshExpireHours int
}

func loadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:             getEnv("JWT_SECRET", ""),
		ExpireHours:        getEnvInt("JWT_EXPIRE_HOURS", 24),
		RefreshExpireHours: getEnvInt("JWT_REFRESH_EXPIRE_HOURS", 168),
	}
}

func (j JWTConfig) AccessTTL() time.Duration  { return time.Duration(j.ExpireHours) * time.Hour }
func (j JWTConfig) RefreshTTL() time.Duration { return time.Duration(j.RefreshExpireHours) * time.Hour }
