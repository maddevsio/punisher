package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
	"github.com/maddevsio/punisher/storage"
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
	gocron.Every(1).Day().At(c.PunishTime).Do(b.checkStandups)

	return b, nil
}

func (b *Bot) sendMsg(update tgbotapi.Update, msg string) {
	text := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
	sm, _ := b.tgAPI.Send(text)
	lastId = sm.MessageID
}

// Start ...
func (b *Bot) Start() {
	go func() {
		<-gocron.Start()
	}()
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
			spew.Dump(update.Message)
			b.db.CreateStandup(model.Standup{
				Comment:  update.Message.Text,
				Username: update.Message.From.UserName,
			})
		}
	}
}
func (b *Bot) checkStandups() {
	lives, err := b.db.ListLives()
	if err != nil {
		fmt.Println(err)
	}
	for _, live := range lives {
		standup, err := b.db.LastStandupFor(live.Username)
		if err != nil {
			fmt.Println(err)
		}
		if time.Now().Day() != standup.Created.Day() {
			live.Lives--
			_, err := b.db.UpdateLive(live)
			if err != nil {
				fmt.Println(err)
			}
			b.LastLives(live)

		}
	}
}
func (b *Bot) isStandup(message *tgbotapi.Message) bool {
	return strings.Contains(message.Text, "#standup")
}

func (b *Bot) LastLives(live model.Live) {
	b.tgAPI.Send(tgbotapi.NewMessage(-1001211952354, fmt.Sprintf("@%s осталось жизней: %d", live.Username, live.Lives)))
}
