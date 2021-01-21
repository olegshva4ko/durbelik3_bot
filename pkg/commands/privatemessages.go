package commands

import (
	"Durbelik3/internal/text"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//PrivateUpdates chan for updates and private processing of messages
var PrivateUpdates chan tgbotapi.Update

//MainPrivate main handler for processing private messages
func (command *CommandAPI) MainPrivate() {
	/*
		Receives updates without Channel posts
	*/
	PrivateUpdates = make(chan tgbotapi.Update)

	for update := range PrivateUpdates {
		if update.CallbackQuery != nil {
			//do some stuff
		}
		if update.Message != nil {
			//means it is global message
			if update.Message.From != nil && update.Message.Chat.ID != int64(update.Message.From.ID) {
				continue
			}

			msgSplitted := strings.Split(update.Message.Text, " ")
			switch msgSplitted[0] {
			case "!NOT READY YET":
			case "!help"://send help message with all stuff
			case "!settings":
			case "settings"://for button
			case "!predict":
			case "!horoscope":
			}

			switch update.Message.Command() {
			case "ping":
				go func() {
					msgToDelete, err := command.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
						text.Ping[command.Buff.Chat[update.Message.Chat.ID].Language]))
					if err != nil {
						return
					}
					time.Sleep(5 * time.Second)
					command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(msgToDelete.Chat.ID, msgToDelete.MessageID))
				}()
			}
		}
	}
}
