package config

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken string `envconfig:"TELEGRAM_TOKEN"`
	DatabaseURL   string `envconfig:"DATABASE_URL"`
	PunishTime    string `envconfig:"PUNISH_TIME" default:"14:00"`
	InternsChatID int64  `envconfig:"INTERNS_CHAT_ID"`
}

// GetConfig ...
func GetConfig() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("bot", &c)
	return &c, err
}
