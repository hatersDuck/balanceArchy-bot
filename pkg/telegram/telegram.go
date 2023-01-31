package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hatersduck/balanceArchy-bot/config"
	"github.com/jackc/pgx"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	messages config.Messages
	db       *pgx.Conn

	add map[int64]*newEvent
}

type newEvent struct {
	main   string
	first  string
	second string
	state  uint8
}

type Answer struct {
	event  string `db:"event"`
	first  string `db:"fir"`
	second string `db:"sec"`
}

func NewBot(bot *tgbotapi.BotAPI, messages config.Messages, conn *pgx.Conn) *Bot {
	return &Bot{
		bot:      bot,
		messages: messages,
		db:       conn,
		add:      make(map[int64]*newEvent),
	}
}

func (b *Bot) Start() error {

	updates, err := b.initUpdatesChanenel()
	if err != nil {
		return err
	}

	b.handlerUpdates(updates)

	return nil
}

func (b *Bot) initUpdatesChanenel() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}
