package telegram_bot

import (
	"fmt"
	"strings"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func LaunchTelegramBot(config configuration_loader.InitialConfiguration, outputChannel chan configuration_loader.Action, exitChannel chan bool) (err error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotConfiguration.TelegramBotToken)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Telegram bot created correctly, waiting for messages")

	actionsMap := make(map[string]messageHandlingFunc)
	actionsMap["start"] = getMessagesAvailableMarkup
	for _, pin := range config.PinsActive {
		actionsMap["turn"+pin.Name+"On"] = turnPinOn
		actionsMap["turn"+pin.Name+"Off"] = turnPinOff
	}

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
			for _, user := range config.TelegramBotConfiguration.AuthorizedUsers {
				if user == update.Message.From.ID {
					userAuthorized = true
				}
			}
			if userAuthorized {
				key := strings.Fields(update.Message.Text)[0]
				if actionFunction, ok := actionsMap[key]; !ok {
					fmt.Println("Action not available")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Action not available")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else {
					msg := actionFunction(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel)
					bot.Send(msg)
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not authorized :(")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
	return nil
}

func getMessagesAvailableMarkup(_ string, config configuration_loader.InitialConfiguration, ChatID int64, ReplyToMessageID int, _ chan configuration_loader.Action) tgbotapi.MessageConfig {
	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	for index, value := range config.PinsActive {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{})
		markup.InlineKeyboard[index] = append(markup.InlineKeyboard[index], tgbotapi.InlineKeyboardButton{Text: "turn" + value.Name + "On"})
		markup.InlineKeyboard[index] = append(markup.InlineKeyboard[index], tgbotapi.InlineKeyboardButton{Text: "turn" + value.Name + "Off"})
	}
	edit := tgbotapi.NewEditMessageReplyMarkup(ChatID, ReplyToMessageID, markup)
	msg := tgbotapi.NewMessage(ChatID, "Action not available")
	msg.ReplyToMessageID = ReplyToMessageID
	msg.ReplyMarkup = edit
	return msg
}

func turnPinOn(message string, config configuration_loader.InitialConfiguration, ChatID int64, ReplyToMessageID int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[4 : len(firstPart)-2]
	outputChannel <- configuration_loader.Action{pin, true}
	return processGPIOResponse(pin, "On", ChatID, ReplyToMessageID)
}

func turnPinOff(message string, config configuration_loader.InitialConfiguration, ChatID int64, ReplyToMessageID int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[4 : len(firstPart)-3]
	outputChannel <- configuration_loader.Action{pin, false}
	return processGPIOResponse(pin, "Off", ChatID, ReplyToMessageID)
}

type messageHandlingFunc func(action string, config configuration_loader.InitialConfiguration, ChatID int64, ReplyToMessageID int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig

func processGPIOResponse(pin string, state string, chatId int64, replyToMessageId int) tgbotapi.MessageConfig {
	response := pin + " turned " + state
	msg := tgbotapi.NewMessage(chatId, response)
	msg.ReplyToMessageID = replyToMessageId
	return msg
}
