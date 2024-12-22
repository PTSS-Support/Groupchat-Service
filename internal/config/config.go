package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	// Firebase Configuration
	FirebaseCredentialFile string `mapstructure:"firebase_credential_file"`

	// Azure Storage Configuration
	AzureStorageAccount   string `mapstructure:"azure_storage_account"`
	AzureStorageKey       string `mapstructure:"azure_storage_key"`
	AzureConnectionString string `mapstructure:"azure_connection_string"`

	// Application Configuration
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	Debug       bool   `mapstructure:"debug"`
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
	viper.BindEnv("azure_storage_account", "AZURE_STORAGE_ACCOUNT")
	viper.BindEnv("azure_storage_key", "AZURE_STORAGE_KEY")
	viper.BindEnv("azure_connection_string", "AZURE_CONNECTION_STRING")
	viper.BindEnv("environment", "APP_ENV")
	viper.BindEnv("port", "APP_PORT")
	viper.BindEnv("debug", "DEBUG")

	// Set defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8080)
	viper.SetDefault("debug", false)

	fmt.Printf("Looking for config file in: %s\n", configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: Config file not found, using environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	fmt.Printf("FirebaseCredentialFile: %s\n", config.FirebaseCredentialFile)
	if config.FirebaseCredentialFile == "" {
		return fmt.Errorf("firebase_credential_file is required")
	}
	if config.AzureStorageAccount == "" {
		return fmt.Errorf("azure_storage_account is required")
	}
	if config.AzureStorageKey == "" {
		return fmt.Errorf("azure_storage_key is required")
	}
	return nil
}
