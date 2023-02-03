package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/hatersduck/balanceArchy-bot/pkg/tdb"
)

const (
	empty = "Неизвестен"
)

func (b *Bot) handlerUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		log.Println(b.add)
		switch {
		case update.Message != nil:
			if update.Message.IsCommand() {
				b.handleCommand(update.Message)
			} else if update.Message.ForwardFrom != nil {
				b.answer(update.Message)
			} else {
				b.handleMessage(update.Message)
			}

		case update.CallbackQuery != nil:
			go b.handleCallback(update.CallbackQuery)

		default:
			log.Printf("Undefiend update %d", update.UpdateID)
		}

	}
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) error {
	id := callback.Message.Chat.ID
	msg := callback.Message

	switch callback.Data {
	case "newEvent":
		b.newEvent(id)

	case "cancel":
		b.add[id] = &Event{}
		b.cancel(id)
		b.comStart(id)
	case "skip_first":
		if val, ok := b.add[id]; ok {
			val.first = empty
			val.state = 2
			msg.Text = empty
			b.handleMessage(msg)
		} else {
			b.comStart(id)
		}

	case "skip_second":
		if val, ok := b.add[id]; ok {
			val.second = empty
			b.Execc(val, id, callback.From.UserName)
		} else {
			b.comStart(id)
		}
	}
	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	id := message.Chat.ID

	if val, ok := b.add[id]; ok {
		switch val.state {
		case 1:
			val.main = message.Text
			val.Answer, _ = tdb.GetEvent(b.db, val.main)
			val.state = 2

			if val.First != empty && val.First != "" {
				mess := tgbotapi.NewMessage(id, "Первый вариант уже создан")
				b.bot.Send(mess)

				message.Text = empty
				b.handleMessage(message)
			} else {
				b.firstEv(id)
			}
		case 2:
			val.first = message.Text
			val.state = 3

			if val.Second != empty && val.Second != "" {
				val.second = empty
				mess := tgbotapi.NewMessage(id, "Второй вариант уже создан")

				b.bot.Send(mess)
				b.Execc(val, id, message.From.UserName)
			} else {
				b.secondEv(id)
			}
		case 3:
			val.second = message.Text
			b.Execc(val, id, message.From.UserName)
		}
	}
	return nil
}

func (b *Bot) Execc(val *Event, id int64, username string) {
	if val.Id == 0 {
		_, err := b.db.Exec("INSERT INTO events (event, fir, sec, user_id, username) VALUES ($1, $2, $3, $4, $5) ",
			val.main, val.first, val.second, id, username)
		if err != nil {
			log.Println("EXEC:", err)
		} else {
			message := tgbotapi.NewMessage(id, "Событие успешно добавлено, cпасибо!")
			b.bot.Send(message)
		}
	} else {
		var err error
		var check bool
		if val.First == empty || val.First == "" {
			_, err = b.db.Exec("UPDATE events SET fir=$1 WHERE id=$2", val.first, val.Id)
			check = true
		}
		if val.Second == empty || val.Second == "" {
			_, err = b.db.Exec("UPDATE events SET sec=$1 WHERE id=$2", val.second, val.Id)
			check = true
		}
		message := tgbotapi.NewMessage(id, "")
		if err != nil {
			message.Text = "Событие не обновлено\n" + err.Error() + "\nОтправьте скриншот диалога @errData"
		} else {
			if check {
				message.Text = "Событие дополнено, спасибо!"
			} else {
				message.Text = "Если вы уверены, что указаны не правильные ответы, то обратитесь к @errData"
			}
		}
		b.bot.Send(message)

	}
	delete(b.add, id)
	b.comStart(id)
}

const (
	commandStart = "/start"
	commandAdd   = "/add"
)

func (b *Bot) handleCommand(command *tgbotapi.Message) error {
	if command.Text == commandStart {
		b.comStart(command.Chat.ID)
	} else if command.Text == commandAdd {
		b.add[command.Chat.ID] = &Event{state: 1}
		b.newEvent(command.Chat.ID)
	}
	return nil
}

func (b *Bot) comStart(id int64) {
	message := tgbotapi.NewMessage(id, b.messages.Start)
	buttons := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.messages.BtnAdd, "newEvent"))
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	b.bot.Send(message)
}

func (b *Bot) newEvent(id int64) {
	b.add[id] = &Event{state: 1}

	message := tgbotapi.NewMessage(id, b.messages.NewEvent)
	buttons := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.messages.BtnCancel, "cancel"))
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	b.bot.Send(message)
}

func (b *Bot) answer(msg *tgbotapi.Message) {
	ans, _ := tdb.GetEvent(b.db, msg.Text)
	var text string
	if ans.Id == 0 {
		text = "Данного события ещё нету вы можете добавить его /add"
	} else {
		text = "Первый вариант: " + ans.First + "\nВторой вариант: " + ans.Second
	}
	mess := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.bot.Send(mess)
}

func (b *Bot) cancel(id int64) {
	message := tgbotapi.NewMessage(id, "Отмена")
	b.bot.Send(message)
}

func (b *Bot) firstEv(id int64) {
	message := tgbotapi.NewMessage(id, b.messages.FirstEvent)
	buttons := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.messages.BtnSkip, "skip_first"))
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	b.bot.Send(message)
}

func (b *Bot) secondEv(id int64) {
	message := tgbotapi.NewMessage(id, b.messages.SecondEvent)
	buttons := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.messages.BtnSkip, "skip_second"))
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	b.bot.Send(message)
}
