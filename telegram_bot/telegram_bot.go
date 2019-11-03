package telegram_bot

import (
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func LaunchTelegramBot(config configuration_loader.TelegramBotConfiguration, outputChannel, inputChannel chan string, exitChannel chan bool) (err error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			select {
			case _ = <-exitChannel:
				return nil
			default:
				continue
			}
		}
		userAuthorized := false
		for _, user := range config.AuthorizedUsers {
			if user == update.Message.From.ID {
				userAuthorized = true
			}
		}
		if userAuthorized {
			outputChannel <- update.Message.Text
			response := <-inputChannel
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not authorized :(")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
	return nil
}
