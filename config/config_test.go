package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const BotToken = "testToken"
const BotChat = "-12345"
const BotDatabaseURL = "root:root@/interns?parseTime=true"

func TestConfig(t *testing.T) {
	os.Setenv("BOT_TELEGRAM_TOKEN", BotToken)
	os.Setenv("BOT_INTERNS_CHAT_ID", BotChat)
	os.Setenv("BOT_DATABASE_URL", BotDatabaseURL)

	c, err := GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "testToken", c.TelegramToken)
	assert.Equal(t, "10:00", c.PunishTime)
	assert.Equal(t, "pushups", c.PunishmentType)

}
