package config

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken  string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	DatabaseURL    string `envconfig:"DATABASE_URL" required:"true"`
	PunishTime     string `envconfig:"PUNISH_TIME" default:"10:00"`
	InternsChatID  int64  `envconfig:"INTERNS_CHAT_ID" required:"true"`
	PunishmentType string `envconfig:"PUNISHMENT_TYPE" default:"pushups"` //also can be "removelives"
}

// GetConfig ...
func GetConfig() (*BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("bot", &c)
	return &c, err
}
