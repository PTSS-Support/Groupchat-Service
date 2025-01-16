package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type Config struct {
	// Firebase Configuration
	FirebaseCredentialFile string `mapstructure:"firebase_credential_file"`

	// Azure Storage Configuration
	AzureConnectionString string `mapstructure:"azure_groupchat_connection_string"`

	// Application Configuration
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	Debug       bool   `mapstructure:"debug"`

	// User Service Configuration
	UserServiceURL        string `mapstructure:"user_service_url"`
	JWKSURL               string `mapstructure:"jwks_url"`
	JWKSJSON              string `mapstructure:"jwks_json"`
	AccessTokenCookieName string `mapstructure:"access_token_cookie_name"`
}

func LoadConfig(configPath string) (*Config, error) {
	var config Config

	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Read environment variables
	viper.AutomaticEnv()

	// Bind environment variables to keys
	viper.BindEnv("firebase_credential_file", "FIREBASE_CREDENTIAL_FILE")
	viper.BindEnv("azure_groupchat_connection_string", "AZURE_GROUPCHAT_CONNECTION_STRING")
	viper.BindEnv("environment", "APP_ENV")
	viper.BindEnv("port", "APP_PORT")
	viper.BindEnv("debug", "DEBUG")
	viper.BindEnv("user_service_url", "USER_SERVICE_URL")
	viper.BindEnv("jwks_url", "JWKS_URL")
	viper.BindEnv("jwks_json", "JWKS_JSON")
	viper.BindEnv("access_token_cookie_name", "ACCESS_TOKEN_COOKIE_NAME")

	// Set defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8080)
	viper.SetDefault("debug", false)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			fmt.Println("Warning: Config file not found, using environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if strings.HasPrefix(config.FirebaseCredentialFile, ".") {
		content, err := os.ReadFile(config.FirebaseCredentialFile)
		if err != nil {
			return nil, fmt.Errorf("error reading firebase credential file: %w", err)
		}
		config.FirebaseCredentialFile = string(content)
	}

	if strings.HasPrefix(config.JWKSJSON, ".") {
		content, err := os.ReadFile(config.JWKSJSON)
		if err != nil {
			return nil, fmt.Errorf("error reading JWKS file: %w", err)
		}
		config.JWKSJSON = string(content)
	}

	if config.JWKSJSON != "" {
		if err := validateJWKSJSON(config.JWKSJSON); err != nil {
			return nil, fmt.Errorf("invalid JWKS JSON: %w", err)
		}
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.AzureConnectionString == "" {
		return fmt.Errorf("azure_connection_string is required")
	}
	if config.FirebaseCredentialFile == "" {
		return fmt.Errorf("firebase_credential_file is required")
	}
	if config.UserServiceURL == "" {
		return fmt.Errorf("user_service_url is required")
	}
	if config.JWKSJSON == "" && config.JWKSURL == "" {
		return fmt.Errorf("either jwks_json or jwks_url is required")
	}
	if config.AccessTokenCookieName == "" {
		return fmt.Errorf("access_token_cookie_name is required")
	}
	return nil
}

func validateJWKSJSON(jwksJSON string) error {
	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.Unmarshal([]byte(jwksJSON), &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS JSON: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return fmt.Errorf("JWKS contains no keys")
	}

	// Validate required fields for at least one signing key
	hasSigningKey := false
	for _, key := range jwks.Keys {
		if key.Use == "sig" && key.Alg == "RS256" {
			if key.Kid == "" || key.N == "" || key.E == "" {
				return fmt.Errorf("signing key missing required fields")
			}
			hasSigningKey = true
			break
		}
	}

	if !hasSigningKey {
		return fmt.Errorf("no valid RS256 signing key found in JWKS")
	}

	return nil
}
