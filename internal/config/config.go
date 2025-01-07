package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	// Firebase Configuration
	FirebaseCredentialFile string `mapstructure:"firebase_credential_file"`

	// Azure Storage Configuration
	AzureConnectionString string `mapstructure:"azure_connection_string"`

	// Application Configuration
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	Debug       bool   `mapstructure:"debug"`

	// User Service Configuration
	UserServiceURL        string `mapstructure:"user_service_url"`
	JWTSecretKey          string `mapstructure:"jwt_secret_key"`
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
	viper.BindEnv("azure_connection_string", "AZURE_CONNECTION_STRING")
	viper.BindEnv("environment", "APP_ENV")
	viper.BindEnv("port", "APP_PORT")
	viper.BindEnv("debug", "DEBUG")
	viper.BindEnv("user_service_url", "USER_SERVICE_URL")
	viper.BindEnv("jwt_secret_key", "JWT_SECRET_KEY")
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
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.FirebaseCredentialFile == "" {
		return fmt.Errorf("firebase_credential_file is required")
	}
	if config.UserServiceURL == "" {
		return fmt.Errorf("user_service_url is required")
	}
	if config.JWTSecretKey == "" {
		return fmt.Errorf("jwt_secret_key is required")
	}
	if config.AccessTokenCookieName == "" {
		return fmt.Errorf("access_token_cookie_name is required")
	}
	return nil
}
