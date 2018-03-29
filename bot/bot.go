package bot

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/maddevsio/telecomedian/config"
	"github.com/maddevsio/telecomedian/model"
	"github.com/maddevsio/telecomedian/storage"
	"github.com/pkg/errors"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	telegramAPIUpdateInterval = 60
	maxResults                = 20
)

var lastId int

// Bot ...
type Bot struct {
	c       *config.BotConfig
	tgAPI   *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	db      *storage.MySQL
}

// NewTGBot creates a new bot
func NewTGBot(c *config.BotConfig) (*Bot, error) {
	newBot, err := tgbotapi.NewBotAPI(c.TelegramToken)
	if err != nil {
		return nil, errors.Wrap(err, "could not create bot")
	}
	b := &Bot{
		c:     c,
		tgAPI: newBot,
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = telegramAPIUpdateInterval
	updates, err := b.tgAPI.GetUpdatesChan(u)
	if err != nil {
		return nil, errors.Wrap(err, "could not create updates chan")
	}
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, errors.Wrap(err, "could not create database connection")
	}
	b.updates = updates
	b.db = conn

	return b, nil
}

func (b *Bot) sendMsg(update tgbotapi.Update, msg string) {
	text := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
	sm, _ := b.tgAPI.Send(text)
	lastId = sm.MessageID
}

// Start ...
func (b *Bot) Start() {
	fmt.Println("Starting tg bot")
	for update := range b.updates {
		if update.Message == nil {
			continue
		}
		spew.Dump(update.EditedMessage)
		text := update.Message.Text
		if text == "" || text == "/start" {
			continue
		}
		if b.isStandup(update.Message) {
			fmt.Printf("accepted standup from %s\n", update.Message.From.UserName)
			b.db.CreateStandup(model.Standup{
				Comment:  update.Message.Text,
				Username: update.Message.From.UserName,
			})
		}
	}
}

func (b *Bot) isStandup(message *tgbotapi.Message) bool {
	return strings.Contains(message.Text, "#standup")
}
