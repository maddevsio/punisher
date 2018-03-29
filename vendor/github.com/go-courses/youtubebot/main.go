package main

import (
	"log"

	"github.com/go-courses/youtubebot/bot"
	"github.com/go-courses/youtubebot/conf"
)

func main() {
	c, err := conf.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	b, err := bot.NewTGBot(c)
	if err != nil {
		log.Fatal(err)
	}
	b.Start()

}
