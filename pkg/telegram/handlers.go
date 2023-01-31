package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	empty = "Неизвестен"
)

func (b *Bot) handlerUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		switch {
		case update.Message != nil:
			if update.Message.IsCommand() {
				b.handleCommand(update.Message)
			} else if update.Message.ForwardFrom != nil {
				if update.Message.ForwardFrom.UserName == "TonarchyBot" {
					b.answer(update.Message)
				}
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
	switch callback.Data {
	case "newEvent":
		b.add[id] = &newEvent{state: 1}
		b.newEvent(id)
	case "cancel":
		b.add[id] = &newEvent{}
		b.cancel(id)
		b.comStart(id)
	case "skip_first":
		b.add[id].first = empty
		b.add[id].state = 3
		b.secondEv(id)
	case "skip_second":
		b.add[id].second = empty
		val := b.add[id]

		b.Execc(val, id, callback.From.UserName)
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	id := message.Chat.ID

	if val, ok := b.add[id]; ok {
		switch val.state {
		case 1:
			val.main = message.Text

			ans := &Answer{}
			err := b.db.QueryRow("SELECT event, fir, sec FROM events WHERE event LIKE $1", val.main+"%").Scan(&ans.event, &ans.first, &ans.second)
			if err != nil {
				log.Println("CHECK:", err)
			}
			if ans.event != "" {
				if ans.first != empty && ans.second != empty {
					mess := tgbotapi.NewMessage(id, "Варианты уже созданы для обновления напишите @errData")
					b.bot.Send(mess)
					b.comStart(id)
				} else if ans.first != empty {
					mess := tgbotapi.NewMessage(id, "Первый вариант уже создан")
					b.bot.Send(mess)
					val.state = 3
					b.secondEv(id)
				}
			} else {
				val.state = 2
				b.firstEv(id)
			}
		case 2:
			val.first = message.Text
			val.state = 3
			ans := &Answer{}
			err := b.db.QueryRow("SELECT event, fir, sec FROM events WHERE event LIKE $1", val.main+"%").Scan(&ans.event, &ans.first, &ans.second)
			if err != nil {
				log.Println("CHECK:", err)
			}
			if ans.event != "" && ans.second != empty {
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

func (b *Bot) Execc(val *newEvent, id int64, username string) {
	ans := &Answer{}
	var sel_id *int
	err := b.db.QueryRow("SELECT id, event, fir, sec FROM events WHERE event LIKE $1", val.main+"%").Scan(&sel_id, &ans.event, &ans.first, &ans.second)
	log.Println(ans)
	if ans.event == "" {
		_, err = b.db.Exec("INSERT INTO events (event, fir, sec, user_id, username) VALUES ($1, $2, $3, $4, $5) ",
			val.main, val.first, val.second, id, username)
		if err != nil {
			log.Println("EXEC:", err)
		} else {
			message := tgbotapi.NewMessage(id, "Событие успешно добавлено, Спасибо!")
			b.bot.Send(message)
		}
	} else {
		if ans.first == empty {
			_, err = b.db.Exec("UPDATE events SET fir=$1 WHERE id=$2", val.first, *sel_id)
		}
		if ans.second == empty {
			_, err = b.db.Exec("UPDATE events SET sec=$1 WHERE id=$2", val.second, *sel_id)
		}
		if err != nil {
			log.Println("UPDATE:", err)
		} else {
			message := tgbotapi.NewMessage(id, "Событие успешно дополнено, Спасибо!")
			b.bot.Send(message)
		}
	}

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
		b.add[command.Chat.ID] = &newEvent{state: 1}
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
	message := tgbotapi.NewMessage(id, b.messages.NewEvent)
	buttons := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.messages.BtnCancel, "cancel"))
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	b.bot.Send(message)
}

func (b *Bot) answer(msg *tgbotapi.Message) {
	ans := &Answer{}
	err := b.db.QueryRow("SELECT fir, ser FROM events WHERE event LIKE $1", msg.Text).Scan(&ans.first, &ans.second)
	if err != nil {
		log.Println("SELECT", err)
	}
	var text string
	if ans.event == "" {
		text = "Данного события ещё нету вы можете добавить его /add"
	} else {
		text = "Первый вариант: " + ans.first + "\nВторой вариант:" + ans.second
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
