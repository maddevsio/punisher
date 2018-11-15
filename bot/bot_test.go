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
const BotDatabaseURL = "root:root@/interns?parseTime=true"

func TestCheckStandups(t *testing.T) {
	b := setupTestBot(t)
	b.dailyJob()
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
	intern, err := b.db.CreateIntern(model.Intern{
		Username: "testUser1",
		Lives:    3,
	})
	assert.NoError(t, err)
	assert.Equal(t, "testUser1", intern.Username)

	message, err := b.checkStandups()
	assert.NoError(t, err)
	assert.Equal(t, "Каратель завершил свою работу ;)", message)

	s, err := b.db.CreateStandup(model.Standup{
		Username: "testUser1",
		Comment:  "first standup",
	})
	assert.NoError(t, err)
	b.dailyJob()

	d = time.Date(2018, time.April, 4, 11, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	b.dailyJob()

	assert.NoError(t, b.db.DeleteIntern(intern.ID))
	assert.NoError(t, b.db.DeleteStandup(s.ID))

}

func TestIsStandup(t *testing.T) {
	b := setupTestBot(t)

	var testCases = []struct {
		message string
		result  bool
	}{
		{"Я написал стэндап!", false},
		{"#standup Я написал стэндап!", false},
		{"#standup", false},
		{"Вчера работал над проектом XYZ, закрыл тикеты 456, 89, 289. Сегодня буду работать над тикетом 203. Проблемы: проект не запускается в докере!", true},
	}

	for _, tt := range testCases {
		isStandup := b.isStandup(&tgbotapi.Message{Text: tt.message})
		assert.Equal(t, tt.result, isStandup)
	}
}

func TestHandleUpdate(t *testing.T) {

	b := setupTestBot(t)
	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{},
	})

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{Text: ""},
	})

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "/start"},
	})

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				UserName: "testUser",
			},
			Text: "Вчера работал. Проблемы: проект не запускается в докере!",
		},
	})

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				UserName: "testUser",
			},
			Text: "@internshipcomedian_bot до @antoliy",
		},
	})

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				UserName: "testUser",
			},
			Text: "@internshipcomedian_bot добавь @antoliyfedorenko",
		},
	})

	interns, err := b.db.ListInterns()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(interns))

	b.handleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{
				UserName: "testUser",
			},
			Text: "@internshipcomedian_bot удали @antoliyfedorenko",
		},
	})

	interns, err = b.db.ListInterns()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(interns))

	standups, err := b.db.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		assert.NoError(t, b.db.DeleteStandup(standup.ID))
	}

}

func TestRemoveLives(t *testing.T) {
	b := setupTestBot(t)
	intern, err := b.db.CreateIntern(model.Intern{
		Username: "testUser1",
		Lives:    3,
	})
	text, err := b.RemoveLives(intern)
	assert.NoError(t, err)
	assert.Equal(t, "@testUser1 осталось жизней: 2", text)
	err = b.db.DeleteIntern(intern.ID)
	assert.NoError(t, err)
}

func TestPunishByPushUps(t *testing.T) {
	b := setupTestBot(t)
	intern, err := b.db.CreateIntern(model.Intern{
		Username: "testUser1",
		Lives:    3,
	})
	pushUps, text, err := b.PunishByPushUps(intern, 0, 10)
	assert.NoError(t, err)
	expected := fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d отжиманий", intern.Username, pushUps)
	assert.Equal(t, expected, text)
	assert.NoError(t, b.db.DeleteIntern(intern.ID))
}

func TestPunishBySitUps(t *testing.T) {
	b := setupTestBot(t)
	intern, err := b.db.CreateIntern(model.Intern{
		Username: "testUser1",
		Lives:    3,
	})
	pushUps, text, err := b.PunishBySitUps(intern, 0, 10)
	assert.NoError(t, err)
	expected := fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d приседаний", intern.Username, pushUps)
	assert.Equal(t, expected, text)
	assert.NoError(t, b.db.DeleteIntern(intern.ID))
}

func TestPunishByPoetry(t *testing.T) {
	b := setupTestBot(t)
	intern, err := b.db.CreateIntern(model.Intern{
		Username: "testUser1",
		Lives:    3,
	})
	l := generatePoetryLink()
	link, text, err := b.PunishByPoetry(intern, l)
	assert.NoError(t, err)
	fmt.Println(link)
	expected := fmt.Sprintf("@%s в наказание за пропущенный стэндап прочитай этот стих: %v", intern.Username, link)
	assert.Equal(t, expected, text)
	assert.NoError(t, b.db.DeleteIntern(intern.ID))
}

func TestPoetryExist(t *testing.T) {
	var testCases = []struct {
		poetrylink string
		result     bool
	}{
		{"https://www.stihi.ru/2011/09/06/21900", false},
		{"https://www.stihi.ru/2011/09/06/219", true},
		{"https://www.stihi.ru/201", false},
		{"", false},
	}

	for _, tt := range testCases {
		ifExist := poetryExist(tt.poetrylink)
		assert.Equal(t, tt.result, ifExist)
	}
}

func TestPunishFunc(t *testing.T) {
	b := setupTestBot(t)
	i, err := b.db.CreateIntern(model.Intern{
		Username: "user",
		Lives:    3,
	})
	assert.NoError(t, err)
	b.Punish(i)
	b.c.PunishmentType = "removelives"
	i2, err := b.db.CreateIntern(model.Intern{
		Username: "user1",
		Lives:    3,
	})
	assert.NoError(t, err)
	b.Punish(i2)
	b.c.PunishmentType = "random"
	b.Punish(i2)
	assert.NoError(t, b.db.DeleteIntern(i.ID))
	assert.NoError(t, b.db.DeleteIntern(i2.ID))

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
