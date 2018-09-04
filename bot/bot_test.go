package bot

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/telegram-bot-api.v4"
)

const BotToken = "testToken"
const BotChat = "-12345"
const BotDatabaseURL = "root:root@/interns"

func TestCheckStandups(t *testing.T) {
	b := setupTestBot(t)
	d := time.Date(2018, time.April, 1, 1, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	_, err := b.checkStandups()
	assert.Equal(t, errors.New("day off").Error(), err.Error())
	d = time.Date(2018, time.April, 7, 1, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	_, err = b.checkStandups()
	assert.Equal(t, errors.New("day off").Error(), err.Error())

	d = time.Date(2018, time.April, 2, 11, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	live, err := b.db.CreateLive(model.Live{
		Username: "testUser1",
		Lives:    3,
	})
	assert.NoError(t, err)
	assert.Equal(t, "testUser1", live.Username)

	message, err := b.checkStandups()
	assert.NoError(t, err)
	assert.Equal(t, "Каратель завершил свою работу ;)", message)

	err = b.db.DeleteLive(live.ID)
	assert.NoError(t, err)

}

func TestIsStandup(t *testing.T) {
	b := setupTestBot(t)

	var testCases = []struct {
		message tgbotapi.Message
		result  bool
	}{
		{tgbotapi.Message{Text: "Я написал стэндап!"}, false},
		{tgbotapi.Message{Text: "#standup Я написал стэндап!"}, true},
	}

	for _, tt := range testCases {
		isStandup := b.isStandup(&tt.message)
		assert.Equal(t, tt.result, isStandup)
	}
}

func TestLastLives(t *testing.T) {
	b := setupTestBot(t)
	live, err := b.db.CreateLive(model.Live{
		Username: "testUser1",
		Lives:    3,
	})
	text, err := b.LastLives(live)
	assert.NoError(t, err)
	assert.Equal(t, "@testUser1 осталось жизней: 3", text)
}

func TestPunishByPushUps(t *testing.T) {
	b := setupTestBot(t)
	live, err := b.db.CreateLive(model.Live{
		Username: "testUser1",
		Lives:    3,
	})
	pushUps, text, err := b.PunishByPushUps(live, 0, 10)
	assert.NoError(t, err)
	expected := fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d отжиманий", live.Username, pushUps)
	assert.Equal(t, expected, text)

}

func setupTestBot(t *testing.T) *Bot {
	os.Setenv("BOT_TELEGRAM_TOKEN", BotToken)
	os.Setenv("BOT_INTERNS_CHAT_ID", BotChat)
	os.Setenv("BOT_DATABASE_URL", BotDatabaseURL)
	conf, err := config.GetConfig()
	assert.NoError(t, err)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	r := httpmock.NewStringResponder(200, `{
		"ok": true,
		"result": {
		  "id": 1000,
		  "is_bot": true,
		  "first_name": "testBot",
		  "username": "testbot_bot"
		}
	  }`)

	url := fmt.Sprintf("https://api.telegram.org/bot%v/getMe", conf.TelegramToken)
	httpmock.RegisterResponder("POST", url, r)

	bot, err := NewTGBot(conf)
	assert.NoError(t, err)
	return bot
}
