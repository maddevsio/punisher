package bot

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/punisher/config"
	"github.com/maddevsio/punisher/model"
	"github.com/maddevsio/punisher/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	telegramAPIUpdateInterval = 60
)

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
		return nil, err
	}
	b := &Bot{
		c:     c,
		tgAPI: newBot,
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = telegramAPIUpdateInterval
	updates, err := b.tgAPI.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	b.updates = updates
	b.db = conn
	gocron.Every(1).Day().At(c.PunishTime).Do(b.dailyJob)

	return b, nil
}

// Start ...
func (b *Bot) Start() {
	go func() {
		<-gocron.Start()
	}()
	logrus.Info("Starting tg bot\n")
	for update := range b.updates {
		b.handleUpdate(update)
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	text := update.Message.Text
	if text == "" || text == "/start" {
		return
	}

	if !strings.Contains(text, "@"+b.tgAPI.Self.UserName) {
		return
	}

	logrus.Infof("New MSG from [%v] chat [%v]\n", update.Message.From.UserName, update.Message.Chat.ID)

	isAdmin, err := b.senderIsAdminInChannel(update.Message.From.UserName, update.Message.Chat.ID)
	if err != nil {
		logrus.Errorf("senderIsAdminInChannel func failed: [%v]\n", err)
	}

	s := strings.Split(update.Message.Text, " ")
	if s[0] == "@"+b.tgAPI.Self.UserName && (s[1] == "добавь" || s[1] == "удали") && s[2] != "" && isAdmin {
		switch c := s[1]; c {
		case "добавь":
			logrus.Infof("Add intern: %s to DB\n", s[2])
			username := strings.Replace(s[2], "@", "", -1)
			intern := model.Intern{0, username, 3}
			_, err := b.db.CreateIntern(intern)
			if err != nil {
				logrus.Errorf("CreateIntern failed: %v", err)
				return
			}
			b.tgAPI.Send(tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s, я слежу за тобой.", intern.Username)))

		case "удали":
			logrus.Infof("Remove intern: %s from DB\n", s[2])
			username := strings.Replace(s[2], "@", "", -1)
			intern, err := b.db.FindIntern(username)
			if err != nil {
				logrus.Errorf("FindIntern failed: %v", err)
				return
			}
			err = b.db.DeleteIntern(intern.ID)
			if err != nil {
				logrus.Errorf("DeleteIntern failed: %v", err)
				return
			}
			b.tgAPI.Send(tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s, я больше не слежу за тобой.", intern.Username)))
		}
	} else {
		if b.isStandup(update.Message) {
			logrus.Infof("accepted standup from %s\n", update.Message.From.UserName)
			standup := model.Standup{
				Comment:  update.Message.Text,
				Username: update.Message.From.UserName,
			}
			_, err := b.db.CreateStandup(standup)
			if err != nil {
				logrus.Errorf("CreateStandup failed: %v\n", err)
				return
			}
			b.tgAPI.Send(tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s спасибо. Я принял твой стендап", update.Message.From.UserName)))

			if b.c.NotifyMentors {
				b.tgAPI.Send(tgbotapi.ForwardConfig{
					FromChannelUsername: update.Message.From.UserName,
					FromChatID:          b.c.InternsChatID,
					MessageID:           update.Message.MessageID,
					BaseChat:            tgbotapi.BaseChat{ChatID: b.c.MentorsChat},
				})
			}

		}
	}

	if update.EditedMessage != nil {

		if !strings.Contains(text, "@"+b.tgAPI.Self.UserName) {
			logrus.Info("MSG does not mention botuser\n")
			return
		}

		if !b.isStandup(update.EditedMessage) {
			logrus.Infof("This is not a proper edit for standup: %s\n", update.EditedMessage.Text)
			return
		}
		logrus.Infof("accepted edited standup from %s\n", update.EditedMessage.From.UserName)
		standup := model.Standup{
			Comment:  update.EditedMessage.Text,
			Username: update.EditedMessage.From.UserName,
		}
		_, err := b.db.UpdateStandup(standup)
		if err != nil {
			logrus.Errorf("UpdateStandup failed: %v\n", err)
			return
		}
		b.tgAPI.Send(tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s спасибо. исправления приняты.", update.Message.From.UserName)))
	}
}

func (b *Bot) dailyJob() {
	if _, err := b.checkStandups(); err != nil {
		logrus.Errorf("checkStandups failed: %v\n", err)
	}
}

func (b *Bot) senderIsAdminInChannel(sendername string, chatID int64) (bool, error) {
	isAdmin := false
	chat := tgbotapi.ChatConfig{chatID, ""}
	admins, err := b.tgAPI.GetChatAdministrators(chat)
	if err != nil {
		return false, err
	}
	for _, admin := range admins {
		if admin.User.UserName == sendername {
			isAdmin = true
			return true, nil
		}
	}
	return isAdmin, nil
}

func (b *Bot) checkStandups() (string, error) {
	logrus.Info("Start checkStandups")
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
			logrus.Info("Intern does not have any standups! Punish")
			if err == sql.ErrNoRows {
				b.Punish(intern)
				continue
			} else {
				return "", err
			}
		}
		t, _ := time.LoadLocation("Asia/Bishkek")
		if time.Now().Day() != standup.Created.In(t).Day() {
			logrus.Infof("Today is %v; last standup created at [%v]", time.Now().Day(), standup.Created.In(t).Day())
			logrus.Info("Intern did not submit standup today! Punish!")
			b.Punish(intern)
		}
	}
	message := tgbotapi.NewMessage(b.c.InternsChatID, "Каратель завершил свою работу ;)")
	b.tgAPI.Send(message)
	return message.Text, nil
}

func (b *Bot) isStandup(message *tgbotapi.Message) bool {
	logrus.Info("checking message...\n")
	var mentionsProblem, mentionsYesterdayWork, mentionsTodayPlans bool

	problemKeys := []string{"роблем", "рудност", "атрдуднен"}
	for _, problem := range problemKeys {
		if strings.Contains(message.Text, problem) {
			mentionsProblem = true
		}
	}

	yesterdayWorkKeys := []string{"чера", "ятницу", "делал", "делано"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message.Text, work) {
			mentionsYesterdayWork = true
		}
	}

	todayPlansKeys := []string{"егодн", "обираюс", "ланир"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message.Text, plan) {
			mentionsTodayPlans = true
		}
	}
	return mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans
}

//RemoveLives removes live from intern
func (b *Bot) RemoveLives(intern model.Intern) (string, error) {
	intern.Lives--
	_, err := b.db.UpdateIntern(intern)
	if err != nil {
		return "", err
	}
	message := tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s осталось жизней: %d", intern.Username, intern.Lives))
	if intern.Lives == 0 {
		chatMemberConf := tgbotapi.ChatMemberConfig{
			ChatID: b.c.InternsChatID,
			UserID: int(intern.ID),
		}
		conf := tgbotapi.KickChatMemberConfig{chatMemberConf, 0}
		r, err := b.tgAPI.KickChatMember(conf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
		message = tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("У @%s не осталось жизней. Удаляю.", intern.Username))
	}
	b.tgAPI.Send(message)
	return message.Text, nil
}

//PunishByPushUps tells interns to do random # of pushups
func (b *Bot) PunishByPushUps(intern model.Intern, min, max int) (int, string, error) {
	rand.Seed(time.Now().Unix())
	pushUps := rand.Intn(max-min) + min
	message := tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d отжиманий", intern.Username, pushUps))
	b.tgAPI.Send(message)
	return pushUps, message.Text, nil
}

//PunishByMakingSnowFlakes tells interns to do random # of pushups
func (b *Bot) PunishByMakingSnowFlakes(intern model.Intern, min, max int) (int, string, error) {
	rand.Seed(time.Now().Unix())
	snowFlakes := rand.Intn(max-min) + min
	message := tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s, в наказание за пропущенный стэндап c тебя %d снежинок!", intern.Username, snowFlakes))
	b.tgAPI.Send(message)
	return snowFlakes, message.Text, nil
}

//PunishBySitUps tells interns to do random # of pushups
func (b *Bot) PunishBySitUps(intern model.Intern, min, max int) (int, string, error) {
	rand.Seed(time.Now().Unix())
	situps := rand.Intn(max-min) + min
	message := tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s в наказание за пропущенный стэндап тебе %d приседаний", intern.Username, situps))
	b.tgAPI.Send(message)
	return situps, message.Text, nil
}

//PunishByPoetry tells interns to read random poetry
func (b *Bot) PunishByPoetry(intern model.Intern, link string) (string, string, error) {
	message := tgbotapi.NewMessage(b.c.InternsChatID, fmt.Sprintf("@%s в наказание за пропущенный стэндап прочитай этот стих на весь офис: %v", intern.Username, link))
	b.tgAPI.Send(message)
	return link, message.Text, nil
}

//Punish punishes interns by either removing lives or asking them to do push ups
func (b *Bot) Punish(intern model.Intern) {
	switch punishment := b.c.PunishmentType; punishment {
	case "pushups":
		b.PunishByPushUps(intern, 5, 100)
	case "snowflakes":
		b.PunishByMakingSnowFlakes(intern, 10, 150)
	case "removelives":
		b.RemoveLives(intern)
	case "situps":
		b.PunishBySitUps(intern, 20, 200)
	case "poetry":
		link := generatePoetryLink()
		b.PunishByPoetry(intern, link)
	case "random":
		b.randomPunishment(intern)
	default:
		b.randomPunishment(intern)
	}
}

func (b *Bot) randomPunishment(intern model.Intern) {
	rand.Seed(time.Now().Unix())
	switch r := rand.Intn(4); r {
	case 0:
		b.PunishByMakingSnowFlakes(intern, 10, 150)
	case 1:
		b.RemoveLives(intern)
	case 2:
		b.PunishBySitUps(intern, 20, 200)
	case 3:
		link := generatePoetryLink()
		b.PunishByPoetry(intern, link)
	case 4:
		b.PunishByPushUps(intern, 5, 100)
	}
}

func generatePoetryLink() string {
	rand.Seed(time.Now().Unix())
	year := rand.Intn(2018-2008) + 2008
	month := rand.Intn(12-1) + 1
	date := rand.Intn(30-1) + 1
	id := rand.Intn(4100-10) + 10
	link := fmt.Sprintf("https://www.stihi.ru/%04d/%02d/%02d/%v", year, month, date, id)
	if !poetryExist(link) {
		generatePoetryLink()
	}
	return link
}

//poetryExist checks if link leads to real poetry and not 404 error
func poetryExist(link string) bool {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		if strings.Contains(bodyString, "404:") {
			return false
		}
		return true
	}

	return false
}
