package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIKey         string
	Port           string
	AWSRegion      string
	IdentityPoolID string
	AWSAccountID   string
	BucketName     string
	PresignExpiry  time.Duration
}

func Load() (*Config, error) {
	expiryMinutes, err := strconv.Atoi(getEnv("PRESIGN_EXPIRY_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid PRESIGN_EXPIRY_MINUTES: %w", err)
	}

	cfg := &Config{
		APIKey:         getEnv("API_KEY", ""),
		Port:           normalizePort(getEnv("PORT", ":8080")),
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		IdentityPoolID: getEnv("IDENTITY_POOL_ID", ""),
		AWSAccountID:   getEnv("AWS_ACCOUNT_ID", ""),
		BucketName:     getEnv("BUCKET_NAME", ""),
		PresignExpiry:  time.Duration(expiryMinutes) * time.Minute,
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY is required")
	}
	if cfg.IdentityPoolID == "" {
		return nil, fmt.Errorf("IDENTITY_POOL_ID is required")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("BUCKET_NAME is required")
	}

	return cfg, nil
}

func normalizePort(port string) string {
	if !strings.HasPrefix(port, ":") {
		return ":" + port
	}
	return port
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
