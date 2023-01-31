package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hatersduck/balanceArchy-bot/config"
	"github.com/hatersduck/balanceArchy-bot/pkg/telegram"
	"github.com/jackc/pgx"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	conn, err := pgx.Connect(pgx.ConnConfig{
		Database: "archydb",
		Password: cfg.DatabasePassword,
		User:     "archy",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	func() {
		if err := telegram.NewBot(bot, cfg.Messages, conn).Start(); err != nil {
			log.Fatal(err)
		}
	}()
}
