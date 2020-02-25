package telegram_bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func LaunchTelegramBot(config configuration_loader.InitialConfiguration, outputChannel chan types.Action, inputChannel chan string, exitChannel chan bool) error {
	bot, err := tgbotapi.NewBotAPI(config.ServerConfiguration.TelegramBotToken)
	if err != nil {
		return err
	}
	fmt.Println("Telegram bot created correctly, waiting for messages")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	go func(updatesChannel tgbotapi.UpdatesChannel) {
		for {
			select {
			case _ = <-exitChannel:
				fmt.Println("Exit signal received in telegram bot")
				exitChannel <- true
				return
			case update := <-updatesChannel:
				if update.Message == nil {
					continue
				}
				userAuthorized := false
				for _, user := range config.ServerConfiguration.TelegramAuthorizedUsers {
					if user == update.Message.From.ID {
						userAuthorized = true
					}
				}
				if userAuthorized {
					messageDivided := strings.Fields(update.Message.Text)
					possibleAction := messageDivided[0]
					if strings.ToLower(possibleAction) == "/start" {
						messages := getMessagesAvailable(outputChannel, inputChannel)
						msg := createMarkupForMessages(messages, update.Message.Chat.ID, update.Message.MessageID)
						bot.Send(msg)
						continue
					} else if matched, err := regexp.Match("OnAndOff$", []byte(possibleAction)); err == nil && matched {
						go func() {
							msg := turnPinOnAndOff(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel, inputChannel)
							bot.Send(msg)
						}()
					} else if matched, err = regexp.Match("On$", []byte(possibleAction)); err == nil && matched {
						go func() {
							msg := turnPinOn(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel, inputChannel)
							bot.Send(msg)
						}()
					} else if matched, err = regexp.Match("Off$", []byte(possibleAction)); err == nil && matched {
						go func() {
							msg := turnPinOff(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel, inputChannel)
							bot.Send(msg)
						}()
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Message was not correct")
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					}
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not authorized :(")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					fmt.Println("User " + strconv.FormatInt(update.Message.Chat.ID, 10) + " tried to send a message (not authorized)")
				}
			}
		}
	}(updates)
	return nil
}

func turnPinOn(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action, inputChannel chan string) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[:len(firstPart)-2]
	outputChannel <- types.Action{pin, true}
	response := <-inputChannel
	return buildMessage(response, chatId, replyToMessageId)
}

func turnPinOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action, inputChannel chan string) tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[:len(firstPart)-3]
	outputChannel <- types.Action{pin, false}
	response := <-inputChannel
	return buildMessage(response, chatId, replyToMessageId)
}

func turnPinOnAndOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action, inputChannel chan string) tgbotapi.MessageConfig {
	fields := strings.Fields(message)
	if len(fields) < 2 {
		return buildMessage("OnAndOff messages should contain at least two words (action and time)", chatId, replyToMessageId)
	}
	firstPart := fields[0]
	pin := firstPart[:len(firstPart)-8]
	duration, err := time.ParseDuration(fields[1])
	if err != nil {
		return buildMessage("Time not set properly", chatId, replyToMessageId)
	}
	outputChannel <- types.Action{pin, true}
	<-inputChannel
	time.Sleep(duration)
	outputChannel <- types.Action{pin, false}
	<-inputChannel
	response := pin + " turned OnAndOff"
	return buildMessage(response, chatId, replyToMessageId)
}

type messageHandlingFunc func(action string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action, inputChannel chan string) tgbotapi.MessageConfig

func buildMessage(msgContent string, chatId int64, replyToMessageId int) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, msgContent)
	msg.ReplyToMessageID = replyToMessageId
	return msg
}

func getMessagesAvailable(outputChannel chan types.Action, inputChannel chan string) []string {
	outputChannel <- types.Action{"start", true}
	message := <-inputChannel
	return strings.Fields(message)
}

func createMarkupForMessages(messages []string, chatId int64, replyToMessageId int) tgbotapi.MessageConfig {
	markup := tgbotapi.NewReplyKeyboard()
	markup.Keyboard = append(markup.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("/start")))
	for _, value := range messages {
		markup.Keyboard = append(markup.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(value+"On"), tgbotapi.NewKeyboardButton(value+"Off"), tgbotapi.NewKeyboardButton(value+"OnAndOff 2s")))
	}
	msg := tgbotapi.NewMessage(chatId, "Welcome to rpi bot")
	msg.ReplyToMessageID = replyToMessageId
	msg.ReplyMarkup = markup
	fmt.Println("User with id \"" + strconv.FormatInt(chatId, 10) + "\" requested message types")
	return msg
}
