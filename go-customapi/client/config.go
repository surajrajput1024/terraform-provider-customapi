package client

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL           string
	AuthURL           string
	Environment       string
	DefaultOrgID      string
	ClientID          string
	Audience          string
	Username          string
	Password          string
	AuthToken         string
}

func LoadConfig() (*Config, error) {
	// Try to load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		BaseURL:      getEnvOrDefault("CUSTOMAPI_BASE_URL"),
		AuthURL:      getEnvOrDefault("CUSTOMAPI_AUTH_URL"),
		Environment:  getEnvOrDefault("CUSTOMAPI_ENVIRONMENT"),
		DefaultOrgID: getEnvOrDefault("CUSTOMAPI_ORG_ID"),
		ClientID:     getEnvOrDefault("CUSTOMAPI_CLIENT_ID"),
		Audience:     getEnvOrDefault("CUSTOMAPI_AUDIENCE"),
		Username:     getEnvOrDefault("CUSTOMAPI_USERNAME"),
		Password:     getEnvOrDefault("CUSTOMAPI_PASSWORD"),
		AuthToken:    getEnvOrDefault("CUSTOMAPI_AUTH_TOKEN"),
	}

	return config, nil
}

func getEnvOrDefault(key string) string {
	return os.Getenv(key)
}


func (c *Config) GetAuthConfig() *AuthConfig {
	return &AuthConfig{
		Username:    c.Username,
		Password:    c.Password,
		AuthToken:   c.AuthToken,
		Environment: c.Environment,
		BaseURL:     c.AuthURL,
	}
}
