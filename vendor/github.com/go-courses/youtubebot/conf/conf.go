package conf

import "github.com/kelseyhightower/envconfig"

// BotConfig ...
type BotConfig struct {
	TelegramToken       string `envconfig:"TELEGRAM_TOKEN"`
	YoutubeDeveloperKey string `envconfig:"YOUTUBE_KEY"`
	WorkingDirectory    string `envconfig:"WORKING_DIRECTORY" default:"files"`
}

// GetConfig ...
func GetConfig() (BotConfig, error) {
	var c BotConfig
	err := envconfig.Process("bot", &c)
	return c, err
}
