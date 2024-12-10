package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration for the specified environment
func Load(env string) (*Config, error) {
	v := viper.New()

	// Load base configuration (local)
	v.SetConfigName("local")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading base config: %v", err)
	}

	// If not local environment, merge with environment specific config
	if env != "local" {
		v.SetConfigName(env)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("error merging %s config: %v", env, err)
		}
	}

	// Enable environment variable substitution
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	return &config, nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
