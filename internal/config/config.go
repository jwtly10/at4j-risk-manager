package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB       PostgresConfig
	Brokers  BrokersConfig
	Jobs     JobsConfig
	Telegram TelegramConfig
	Port     string
	ApiKey   string
}

type JobsConfig struct {
	// Interval in seconds to check equity
	EquityCheckInterval int
}

type PostgresConfig struct {
	Username string
	Password string
	URL      string
	Port     string
	DBName   string
}

type TelegramConfig struct {
	Token  string
	ChatId string
}

type BrokersConfig struct {
	Oanda OandaConfig
	MT5   MT5Config
}

type OandaConfig struct {
	ApiKey  string
	BaseUrl string
}

type MT5Config struct {
	ApiKey  string
	BaseUrl string
}

func LoadConfig() (*Config, error) {
	// We validate the environment variables anyway
	_ = godotenv.Load()

	cfg := &Config{}

	if os.Getenv("PORT") == "" {
		cfg.Port = "8001"
	} else {
		cfg.Port = os.Getenv("PORT")
	}

	if os.Getenv("INTERNAL_API_KEY") == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	} else {
		cfg.ApiKey = os.Getenv("INTERNAL_API_KEY")
	}

	eqInt, err := strconv.Atoi(os.Getenv("EQUITY_CHECK_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse EQUITY_CHECK_INTERVAL: %v", err)
	}
	cfg.Jobs = JobsConfig{
		EquityCheckInterval: eqInt,
	}

	cfg.DB = PostgresConfig{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		URL:      os.Getenv("DB_URL"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
	}

	o := OandaConfig{
		ApiKey:  os.Getenv("OANDA_API_KEY"),
		BaseUrl: os.Getenv("OANDA_API_URL"),
	}

	m := MT5Config{
		ApiKey:  os.Getenv("MT5_API_KEY"),
		BaseUrl: os.Getenv("MT5_API_URL"),
	}

	t := TelegramConfig{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatId: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	cfg.Telegram = t

	cfg.Brokers = BrokersConfig{
		Oanda: o,
		MT5:   m,
	}

	if err := cfg.DB.validate(); err != nil {
		return nil, err
	}

	if err := cfg.Brokers.validate(); err != nil {
		return nil, err
	}

	if err := cfg.Jobs.validate(); err != nil {
		return nil, err
	}

	if err := cfg.Telegram.validate(); err != nil {
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

func (b BrokersConfig) validate() error {
	// Validate Oanda configuration
	if b.Oanda.ApiKey == "" {
		return fmt.Errorf("OANDA_API_KEY is required")
	}
	if b.Oanda.BaseUrl == "" {
		return fmt.Errorf("OANDA_API_URL is required")
	}

	// Validate MT5 configuration
	if b.MT5.ApiKey == "" {
		return fmt.Errorf("MT5_API_KEY is required")
	}
	if b.MT5.BaseUrl == "" {
		return fmt.Errorf("MT5_API_URL is required")
	}

	return nil
}

func (j JobsConfig) validate() error {
	if j.EquityCheckInterval == 0 {
		return fmt.Errorf("EQUITY_CHECK_INTERVAL is required and CANNOT be 0")
	}

	return nil
}

func (t TelegramConfig) validate() error {
	if t.Token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if t.ChatId == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	return nil
}
