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

	// Service Discovery Configuration
	ServiceDiscovery ServiceDiscoveryConfig `mapstructure:"service_discovery"`

	// User Service Configuration
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
	viper.BindEnv("jwt_secret_key", "JWT_SECRET_KEY")
	viper.BindEnv("access_token_cookie_name", "ACCESS_TOKEN_COOKIE_NAME")

	// Service Discovery environment bindings
	viper.BindEnv("service_discovery.namespace", "SERVICE_DISCOVERY_NAMESPACE")
	viper.BindEnv("service_discovery.service_name", "SERVICE_DISCOVERY_SERVICE_NAME")
	viper.BindEnv("service_discovery.port_name", "SERVICE_DISCOVERY_PORT_NAME")
	viper.BindEnv("service_discovery.refresh_interval", "SERVICE_DISCOVERY_REFRESH_INTERVAL")

	// Set defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8080)
	viper.SetDefault("debug", false)
	viper.SetDefault("service_discovery.refresh_interval", 30)

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
	if config.ServiceDiscovery.Namespace == "" {
		return fmt.Errorf("service_discovery.namespace is required")
	}
	if config.ServiceDiscovery.ServiceName == "" {
		return fmt.Errorf("service_discovery.service_name is required")
	}
	if config.ServiceDiscovery.PortName == "" {
		return fmt.Errorf("service_discovery.port_name is required")
	}
	if config.JWTSecretKey == "" {
		return fmt.Errorf("jwt_secret_key is required")
	}
	if config.AccessTokenCookieName == "" {
		return fmt.Errorf("access_token_cookie_name is required")
	}
	return nil
}
