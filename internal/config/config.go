package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string   `mapstructure:"port"`
	DatabaseURL    string   `mapstructure:"database_url"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	JWTSecret      string   `mapstructure:"jwt_secret"`
	FCMServerKey   string   `mapstructure:"fcm_server_key"`
}

func LoadConfig() (*Config, error) {
	// Set configuration file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add paths to look for the config file
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/fcm-microservice")

	// Read environment variables
	viper.AutomaticEnv()

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %s", err)
	}

	// Override with environment variables if set
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	return &config, nil
}
