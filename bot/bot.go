package bot

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

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
	minPushUps                = 50
	maxPushUps                = 500
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
	gocron.Every(1).Day().At(c.PunishTime).Do(b.dailyJob)

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
	log.Println("Starting tg bot")
	for update := range b.updates {
		if update.Message == nil {
			continue
		}
		text := update.Message.Text
		if text == "" || text == "/start" {
			continue
		}
		if b.isStandup(update.Message) {
			fmt.Printf("accepted standup from %s\n", update.Message.From.UserName)
			if _, err := b.db.CreateStandup(model.Standup{
				Comment:  update.Message.Text,
				Username: update.Message.From.UserName,
			}); err != nil {
				log.Println(err)
				continue
			}
			b.tgAPI.Send(tgbotapi.NewMessage(-b.c.InternsChatID, fmt.Sprintf("@%s спасибо. Я принял твой стендап", update.Message.From.UserName)))

			b.tgAPI.Send(tgbotapi.ForwardConfig{
				FromChannelUsername: update.Message.From.UserName,
				FromChatID:          -b.c.InternsChatID,
				MessageID:           update.Message.MessageID,
				BaseChat:            tgbotapi.BaseChat{ChatID: -319163668},
			})

		}
	}
}

func (b *Bot) dailyJob() {
	if _, err := b.checkStandups(); err != nil {
		log.Println(err)
	}
}
func (b *Bot) checkStandups() (string, error) {
	if time.Now().Weekday().String() == "Saturday" || time.Now().Weekday().String() == "Sunday" {
		return "", errors.New("day off")
	}
	interns, err := b.db.ListInterns()
	if err != nil {
		return "", err
	}
	for _, intern := range interns {
		standup, err := b.db.LastStandupFor(intern.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				b.Punish(intern)
				continue
			}
		}
		t, err := time.LoadLocation("Asia/Bishkek")
		if err != nil {
			log.Println(err)
		}
		if time.Now().Day() != standup.Created.In(t).Day() {
			b.Punish(intern)
		}
	}
	message := tgbotapi.NewMessage(-b.c.InternsChatID, "Каратель завершил свою работу ;)")
	b.tgAPI.Send(message)
	return message.Text, nil
}

func (b *Bot) isStandup(message *tgbotapi.Message) bool {
	log.Println("checking accepted message")
	mentionsProblem := false
	problemKeys := []string{"роблем", "рудност", "атрдуднен"}
	for _, problem := range problemKeys {
		if strings.Contains(message.Text, problem) {
			mentionsProblem = true
		}
	}

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{"чера", "ятницу", "делал", "делано"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message.Text, work) {
			mentionsYesterdayWork = true
		}
	}
	mentionsTodayPlans := false
	todayPlansKeys := []string{"егодн", "обираюс", "ланир"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message.Text, plan) {
			mentionsTodayPlans = true
		}
	}
	if mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans {
		return true
	}
	return false
}

//RemoveLives removes live from intern
func (b *Bot) RemoveLives(intern model.Intern) (string, error) {
	intern.Lives--
	_, err := b.db.UpdateIntern(intern)
	if err != nil {
		log.Println(err)
	}
	message := tgbotapi.NewMessage(-b.c.InternsChatID, fmt.Sprintf("@%s осталось жизней: %d", intern.Username, intern.Lives))
	b.tgAPI.Send(message)
	return message.Text, nil
}

//PunishByPushUps tells interns to do random # of pushups
func (b *Bot) PunishByPushUps(intern model.Intern, min, max int) (int, string, error) {
	rand.Seed(time.Now().Unix())
	pushUps := rand.Intn(max-min) + min
	message := tgbotapi.NewMessage(-b.c.InternsChatID, fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d отжиманий", intern.Username, pushUps))
	b.tgAPI.Send(message)
	return pushUps, message.Text, nil
}

//Punish punishes interns by either removing lives or asking them to do push ups
func (b *Bot) Punish(intern model.Intern) {
	b.RemoveLives(intern)
	b.PunishByPushUps(intern, minPushUps, maxPushUps)
}
