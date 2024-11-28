package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB      PostgresConfig
	Brokers BrokersConfig
	Port    string
}

type PostgresConfig struct {
	Username string
	Password string
	URL      string
	Port     string
	DBName   string
}

// TODO: How should we handle different brokers (oanda/mt5 adapter service)
type BrokersConfig struct {
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	cfg := &Config{}

	if os.Getenv("PORT") == "" {
		cfg.Port = "8001"
	}

	cfg.DB = PostgresConfig{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		URL:      os.Getenv("DB_URL"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
	}

	cfg.Brokers = BrokersConfig{}

	if err := cfg.DB.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if the required environment variables are set
func (p PostgresConfig) validate() error {
	required := map[string]string{
		"DB_USERNAME": p.Username,
		"DB_PASSWORD": p.Password,
		"DB_URL":      p.URL,
		"DB_PORT":     p.Port,
	}

	for env, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required", env)
		}
	}

	return nil
}
