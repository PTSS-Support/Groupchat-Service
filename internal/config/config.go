package config

import (
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
	PublicKey             string `mapstructure:"keycloak_public_key"`
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
	viper.BindEnv("keycloak_public_key", "KEYCLOAK_PUBLIC_KEY")
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
	if config.PublicKey == "" {
		return fmt.Errorf("keycloak_public_key is required")
	}
	if config.AccessTokenCookieName == "" {
		return fmt.Errorf("access_token_cookie_name is required")
	}
	return nil
}
