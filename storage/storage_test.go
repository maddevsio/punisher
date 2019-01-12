package storage

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
	"github.com/stretchr/testify/assert"
)

const BotToken = "testToken"
const BotChat = "-12345"
const BotDatabaseURL = "root:root@/interns?parseTime=true"

func TestCRUDLStandup(t *testing.T) {
	m, err := setup()
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

func TestInternFunctionality(t *testing.T) {
	m, err := setup()
	assert.NoError(t, err)

	i, err := m.CreateIntern(model.Intern{
		Username: "user",
		Lives:    3,
	})
	assert.NoError(t, err)

	intern, err := m.SelectIntern(i.ID)
	assert.NoError(t, err)
	assert.Equal(t, "user", intern.Username)
	assert.Equal(t, 3, intern.Lives)

	s, err := m.CreateStandup(model.Standup{
		Comment:  "work hard",
		Username: "user",
	})
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)

	s1, err := m.CreateStandup(model.Standup{
		Comment:  "work very hard after 1 sec",
		Username: "user",
	})
	assert.NoError(t, err)

	lastStandup, err := m.LastStandupFor("user")
	assert.NoError(t, err)
	assert.Equal(t, "work very hard after 1 sec", lastStandup.Comment)

	i1, err := m.CreateIntern(model.Intern{
		Username: "user2",
		Lives:    1,
	})
	assert.NoError(t, err)

	i1.Username = "newuser2"
	updatedIntern, err := m.UpdateIntern(i1)
	assert.NoError(t, err)
	assert.Equal(t, "newuser2", updatedIntern.Username)

	interns, err := m.ListInterns()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(interns))

	assert.NoError(t, m.DeleteStandup(s.ID))
	assert.NoError(t, m.DeleteStandup(s1.ID))
	assert.NoError(t, m.DeleteIntern(i.ID))
	assert.NoError(t, m.DeleteIntern(i1.ID))
}

func setup() (*MySQL, error) {
	os.Setenv("BOT_TELEGRAM_TOKEN", BotToken)
	os.Setenv("BOT_INTERNS_CHAT_ID", BotChat)
	os.Setenv("BOT_DATABASE_URL", BotDatabaseURL)
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	m, err := NewMySQL(c)
	if err != nil {
		return nil, err
	}
	return m, nil
}
