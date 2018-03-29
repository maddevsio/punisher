package bot

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-courses/youtubebot/conf"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	telegramAPIUpdateInterval = 60
	maxResults                = 20
)

var lastId int

// Bot ...
type Bot struct {
	c       conf.BotConfig
	tgAPI   *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	yClient *youtube.Service
}

// NewTGBot creates a new bot
func NewTGBot(c conf.BotConfig) (*Bot, error) {
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
	b.updates = updates
	client, err := youtube.New(&http.Client{
		Transport: &transport.APIKey{Key: c.YoutubeDeveloperKey},
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Key ", c.YoutubeDeveloperKey)
	b.yClient = client
	if _, err := os.Stat(c.WorkingDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(c.WorkingDirectory, os.ModePerm); err != nil {
			log.Println(err)
		}
	}
	return b, nil
}

func (b *Bot) sendMsg(update tgbotapi.Update, msg string) {
	text := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
	sm, _ := b.tgAPI.Send(text)
	lastId = sm.MessageID
}
func (b *Bot) sendAudio(update tgbotapi.Update, filePath string) {
	audio := tgbotapi.NewAudioUpload(update.Message.Chat.ID, filePath)
	b.tgAPI.Send(audio)
}

func (b *Bot) search(searchText string) (string, error) {
	// Make the API call to YouTube.
	call := b.yClient.Search.List("id,snippet").
		Q(searchText).MaxResults(maxResults)
	response, err := call.Do()
	if err != nil {
		return "", errors.Wrap(err, "could not find videos on youtube")
	}
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			return item.Id.VideoId, nil
		}
	}

	return "", errors.New("unknown error for youtube")
}

// Start ...
func (b *Bot) Start() {
	fmt.Println("Starting tg bot")
	for update := range b.updates {
		if update.Message == nil {
			continue
		}
		text := update.Message.Text
		if text == "" || text == "/start" {
			continue
		}

		converted := make(chan bool, 1)
		searched := make(chan bool, 1)
		b.sendMsg(update, "Начал поиск")
		fmt.Println(lastId)

		go func() {
			for {
				tex := tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Ищу.")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Ищу..")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Ищу...")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Ищу....")
				b.tgAPI.Send(tex)
				if <-searched {
					break
				}
			}
		}()

		youtubeID, err := b.search(text)
		if err != nil {
			log.Println("could not get video id from youtube", err)
		}
		url, title, err := GetDownloadURL(youtubeID)
		if err != nil {
			log.Println("could not get download url", err)
		}
		searched <- true

		go func() {
			for {
				tex := tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Конвертирую.")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Конвертирую..")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Конвертирую...")
				b.tgAPI.Send(tex)
				tex = tgbotapi.NewEditMessageText(update.Message.Chat.ID, lastId, "Конвертирую....")
				b.tgAPI.Send(tex)
				if <-converted {
					break
				}
			}
		}()
		err = Convert(title, url)
		if err != nil {
			log.Println("could not convert video file to mp3 ", err)
		}
		converted <- true
		fileName := fmt.Sprintf("%s/%s.mp3", b.c.WorkingDirectory, title)
		b.sendAudio(update, fileName)
		os.Remove(fileName)
	}
}
