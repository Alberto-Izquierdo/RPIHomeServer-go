package telegram_bot

import (
	"fmt"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func LaunchTelegramBot(config configuration_loader.TelegramBotConfiguration, outputChannel, inputChannel chan string, exitChannel chan bool) (err error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Telegram bot created correctly, waiting for messages")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for {
		select {
		case _ = <-exitChannel:
			return nil
		case update := <-updates:
			if update.Message == nil {
				select {
				case _ = <-exitChannel:
					fmt.Println("Exit signal received, exiting from telegram bot update")
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
	}
	return nil
}
