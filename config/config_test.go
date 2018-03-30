package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "value")

	c, err := GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "value", c.TelegramToken)
	assert.Equal(t, "14:00", c.PunishTime)

}
