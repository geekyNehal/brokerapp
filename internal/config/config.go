package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Database Configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT Configuration
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	// Server Configuration
	ServerPort string

	// Database Pool Configuration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (*Config, error) {
	// Parse durations
	accessTokenDuration, err := time.ParseDuration(getEnv("ACCESS_TOKEN_DURATION", "5m"))
	if err != nil {
		return nil, fmt.Errorf("Invalid ACCESS_TOKEN_DURATION: %v", err)
	}

	refreshTokenDuration, err := time.ParseDuration(getEnv("REFRESH_TOKEN_DURATION", "24h"))
	if err != nil {
		return nil, fmt.Errorf("Invalid REFRESH_TOKEN_DURATION: %v", err)
	}

	connMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("Invalid DB_CONN_MAX_LIFETIME: %v", err)
	}

	// Parse integers
	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("Invalid DB_MAX_OPEN_CONNS: %v", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("Invalid DB_MAX_IDLE_CONNS: %v", err)
	}

	cfg := &Config{
		// Database Configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "brokerapp"),

		// JWT Configuration
		JWTSecret:            getEnv("JWT_SECRET"),
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,

		// Server Configuration
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Database Pool Configuration
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}

	// Validate required environment variables
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Print configuration (without sensitive data)
	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("DB_HOST: %s\n", cfg.DBHost)
	fmt.Printf("DB_PORT: %s\n", cfg.DBPort)
	fmt.Printf("DB_USER: %s\n", cfg.DBUser)
	fmt.Printf("DB_NAME: %s\n", cfg.DBName)
	fmt.Printf("JWT_SECRET: %s...\n", cfg.JWTSecret[:5])
	fmt.Printf("SERVER_PORT: %s\n", cfg.ServerPort)
	fmt.Printf("ACCESS_TOKEN_DURATION: %v\n", cfg.AccessTokenDuration)
	fmt.Printf("REFRESH_TOKEN_DURATION: %v\n", cfg.RefreshTokenDuration)
	fmt.Printf("DB_MAX_OPEN_CONNS: %d\n", cfg.MaxOpenConns)
	fmt.Printf("DB_MAX_IDLE_CONNS: %d\n", cfg.MaxIdleConns)
	fmt.Printf("DB_CONN_MAX_LIFETIME: %v\n", cfg.ConnMaxLifetime)

	return cfg, nil
}

func getEnv(key string, defaultValue ...string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}
