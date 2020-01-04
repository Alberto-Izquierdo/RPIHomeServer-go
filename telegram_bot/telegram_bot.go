package telegram_bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
		actionsMap["turn"+pin.Name+"OnAndOff"] = turnPinOnAndOff
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
					fmt.Println("Action \"" + key + "\" not available")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Action \""+key+"\" not available")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else {
					go func() {
						msg := actionFunction(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel)
						bot.Send(msg)
					}()
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not authorized :(")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				fmt.Println("User " + strconv.FormatInt(update.Message.Chat.ID, 10) + " tried to send a message (not authorized)")
			}
		}
	}
	return nil
}

func getMessagesAvailableMarkup(_ string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, _ chan configuration_loader.Action) tgbotapi.MessageConfig {
	markup := tgbotapi.NewReplyKeyboard()
	for _, value := range config.PinsActive {
		markup.Keyboard = append(markup.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("turn"+value.Name+"On"), tgbotapi.NewKeyboardButton("turn"+value.Name+"Off"), tgbotapi.NewKeyboardButton("turn"+value.Name+"OnAndOff 2s")))
	}
	msg := tgbotapi.NewMessage(chatId, "Welcome to rpi bot")
	msg.ReplyToMessageID = replyToMessageId
	msg.ReplyMarkup = markup
	fmt.Println("User with id \"" + strconv.FormatInt(chatId, 10) + "\" requested message types")
	return msg
}

func turnPinOn(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[4 : len(firstPart)-2]
	outputChannel <- configuration_loader.Action{pin, true}
	response := pin + " turned On"
	return buildMessage(response, chatId, replyToMessageId)
}

func turnPinOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[4 : len(firstPart)-3]
	outputChannel <- configuration_loader.Action{pin, false}
	response := pin + " turned Off"
	return buildMessage(response, chatId, replyToMessageId)
}

func turnPinOnAndOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig {
	fields := strings.Fields(message)
	if len(fields) < 2 {
		return buildMessage("OnAndOff messages should contain at least two words (action and time)", chatId, replyToMessageId)
	}
	firstPart := fields[0]
	pin := firstPart[4 : len(firstPart)-8]
	duration, err := time.ParseDuration(fields[1])
	if err != nil {
		return buildMessage("Time not set properly", chatId, replyToMessageId)
	}
	outputChannel <- configuration_loader.Action{pin, true}
	time.Sleep(duration)
	outputChannel <- configuration_loader.Action{pin, false}
	response := pin + " turned OnAndOff"
	return buildMessage(response, chatId, replyToMessageId)
}

type messageHandlingFunc func(action string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan configuration_loader.Action) tgbotapi.MessageConfig

func buildMessage(msgContent string, chatId int64, replyToMessageId int) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, msgContent)
	msg.ReplyToMessageID = replyToMessageId
	return msg
}
