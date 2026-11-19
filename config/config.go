package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App  AppConfig
	DB   DBConfig
	JWT  JWTConfig
	CORS CORSConfig
}

type AppConfig struct {
	Env  string
	Port string
}

type DBConfig struct {
	MasterDSN   string
	ReplicaDSN  string
	MaxConns    int32
	MinConns    int32
	MaxConnIdle time.Duration
	MaxConnLife time.Duration
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("DB_MASTER_DSN", "postgres://postgres:postgres@localhost:5432/loyalty_databases?sslmode=disable")
	v.SetDefault("DB_REPLICA_DSN", "postgres://postgres:postgres@localhost:5432/loyalty_databases?sslmode=disable")
	v.SetDefault("DB_MAX_CONNS", 50)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DB_MAX_CONN_IDLE", "5m")
	v.SetDefault("DB_MAX_CONN_LIFE", "30m")
	v.SetDefault("JWT_SECRET", "super-secret-loyalty-backoffice-jwt-key-2024")
	v.SetDefault("JWT_ACCESS_EXPIRY", "60m")
	v.SetDefault("JWT_REFRESH_EXPIRY", "168h") // 7 days
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")

	v.AutomaticEnv()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // ignore error if .env doesn't exist

	idleDuration, _ := time.ParseDuration(v.GetString("DB_MAX_CONN_IDLE"))
	if idleDuration == 0 {
		idleDuration = 5 * time.Minute
	}
	lifeDuration, _ := time.ParseDuration(v.GetString("DB_MAX_CONN_LIFE"))
	if lifeDuration == 0 {
		lifeDuration = 30 * time.Minute
	}
	accessExpiry, _ := time.ParseDuration(v.GetString("JWT_ACCESS_EXPIRY"))
	if accessExpiry == 0 {
		accessExpiry = 60 * time.Minute
	}
	refreshExpiry, _ := time.ParseDuration(v.GetString("JWT_REFRESH_EXPIRY"))
	if refreshExpiry == 0 {
		refreshExpiry = 168 * time.Hour
	}

	originsRaw := v.GetString("CORS_ALLOWED_ORIGINS")
	origins := strings.Split(originsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("APP_ENV"),
			Port: v.GetString("APP_PORT"),
		},
		DB: DBConfig{
			MasterDSN:   v.GetString("DB_MASTER_DSN"),
			ReplicaDSN:  v.GetString("DB_REPLICA_DSN"),
			MaxConns:    int32(v.GetInt("DB_MAX_CONNS")),
			MinConns:    int32(v.GetInt("DB_MIN_CONNS")),
			MaxConnIdle: idleDuration,
			MaxConnLife: lifeDuration,
		},
		JWT: JWTConfig{
			Secret:        v.GetString("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		CORS: CORSConfig{
			AllowedOrigins: origins,
		},
	}

	return cfg, nil
}

