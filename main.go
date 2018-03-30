package main

import (
	"log"

	"github.com/maddevsio/telecomedian/bot"
	"github.com/maddevsio/telecomedian/config"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	b, err := bot.NewTGBot(c)
	if err != nil {
		log.Fatal(err)
	}

	b.Start()
}
