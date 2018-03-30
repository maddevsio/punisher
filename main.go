package main

import (
	"log"

	"github.com/maddevsio/punisher/bot"
	"github.com/maddevsio/punisher/config"
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
