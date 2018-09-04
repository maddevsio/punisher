package storage

import (
	"database/sql"
	"os"
	"testing"

	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
	"github.com/stretchr/testify/assert"
)

const BotToken = "testToken"
const BotChat = "-12345"
const BotDatabaseURL = "root:root@/interns?parseTime=true"

func TestCRUDLStandup(t *testing.T) {
	os.Setenv("BOT_TELEGRAM_TOKEN", BotToken)
	os.Setenv("BOT_INTERNS_CHAT_ID", BotChat)
	os.Setenv("BOT_DATABASE_URL", BotDatabaseURL)

	c, err := config.GetConfig()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	s, err := m.CreateStandup(model.Standup{
		Comment:  "work hard",
		Username: "user",
	})
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "work hard")
	s.Comment = "Rest"
	s, err = m.UpdateStandup(s)
	assert.NoError(t, err)
	assert.Equal(t, "Rest", s.Comment)
	items, err := m.ListStandups()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(items))
	selected, err := m.SelectStandup(s.ID)
	assert.NoError(t, err)
	assert.Equal(t, s, selected)
	assert.NoError(t, m.DeleteStandup(s.ID))
	s, err = m.SelectStandup(s.ID)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Equal(t, s.ID, int64(0))

}
