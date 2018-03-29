package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Setenv("TELEGRAM_TOKEN", "value")
	os.Setenv("YOUTUBE_KEY", "key")
	c, err := GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "value", c.TelegramToken)
	assert.Equal(t, "key", c.YoutubeDeveloperKey)

}
