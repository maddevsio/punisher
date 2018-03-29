package config

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken string `envconfig:"TELEGRAM_TOKEN"`
	DatabaseURL   string `envconfig:"DATABASE_URL"`
}

// GetConfig ...
func GetConfig() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("bot", &c)
	return &c, err
}
