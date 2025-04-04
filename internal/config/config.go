package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// Config stores all application configurations
type Config struct {
	LogLevel string `mapstructure:"log_level" envconfig:"LOG_LEVEL" default:"development"`
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

// ServerConfig stores HTTP server configurations
type ServerConfig struct {
	Address string `mapstructure:"address" envconfig:"SERVER_ADDRESS" default:":8080"`
}

// DatabaseConfig stores PostgreSQL configurations
type DatabaseConfig struct {
	Host     string `mapstructure:"host" envconfig:"DB_HOST" default:"localhost"`
	Port     int    `mapstructure:"port" envconfig:"DB_PORT" default:"5432"`
	User     string `mapstructure:"user" envconfig:"DB_USER" default:"postgres"`
	Password string `mapstructure:"password" envconfig:"DB_PASSWORD" default:"postgres"`
	DBName   string `mapstructure:"dbname" envconfig:"DB_NAME" default:"app_db"`
	SSLMode  string `mapstructure:"sslmode" envconfig:"DB_SSLMODE" default:"disable"`
}

// DSN returns the PostgreSQL connection string
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// RedisConfig stores Redis configurations
type RedisConfig struct {
	Address  string `mapstructure:"address" envconfig:"REDIS_ADDRESS" default:"localhost:6379"`
	Password string `mapstructure:"password" envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `mapstructure:"db" envconfig:"REDIS_DB" default:"0"`
}

// Load loads configurations from files and environment variables
func Load() (*Config, error) {
	var config Config

	// Set up Viper configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading configuration file: %w", err)
		}
		// Do not return an error if the configuration file is missing; use defaults and environment variables
	}

	// Load environment-specific configurations
	env := os.Getenv("APP_ENV")
	if env != "" {
		viper.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading environment-specific configuration file: %w", err)
			}
		}
	}

	// Unmarshal configuration into the struct
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	// Load configuration from environment variables
	if err := envconfig.Process("app", &config); err != nil {
		return nil, fmt.Errorf("error processing environment variables: %w", err)
	}

	return &config, nil
}
