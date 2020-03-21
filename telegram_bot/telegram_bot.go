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

func LaunchTelegramBot(config configuration_loader.InitialConfiguration, outputChannel chan types.Action, programmedActionOperationsChannel chan types.ProgrammedActionOperation, inputChannel chan types.TelegramMessage, exitChannel chan bool) error {
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
			createProgrammedActionRegex := regexp.MustCompile("^CreateProgrammedAction (.*)$")
			removeProgrammedActionRegex := regexp.MustCompile("^RemoveProgrammedAction (.*)$")
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
						outputChannel <- types.Action{"start", true, update.Message.Chat.ID}
						continue
					} else if matched, err := regexp.Match("OnAndOff$", []byte(possibleAction)); err == nil && matched {
						go func() {
							msg := turnPinOnAndOff(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel)
							if msg != nil {
								bot.Send(msg)
							}
						}()
					} else if matched, err = regexp.Match("On$", []byte(possibleAction)); err == nil && matched {
						go turnPinOn(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel)
					} else if matched, err = regexp.Match("Off$", []byte(possibleAction)); err == nil && matched {
						go turnPinOff(update.Message.Text, config, update.Message.Chat.ID, update.Message.MessageID, outputChannel)
					} else if matchedGroups := createProgrammedActionRegex.FindStringSubmatch(update.Message.Text); len(matchedGroups) > 1 {
						go func() {
							msg := createProgrammedAction(matchedGroups[1], update.Message.Chat.ID, programmedActionOperationsChannel)
							if msg != nil {
								bot.Send(msg)
							}
						}()
					} else if matchedGroups := removeProgrammedActionRegex.FindStringSubmatch(update.Message.Text); len(matchedGroups) > 1 {
						go func() {
							msg := removeProgrammedAction(matchedGroups[1], update.Message.Chat.ID, programmedActionOperationsChannel)
							if msg != nil {
								bot.Send(msg)
							}
						}()
					} else if matched, err = regexp.Match("^GetProgrammedActions$", []byte(possibleAction)); err == nil && matched {
						// TODO: return programmed actions and keyboard to remove them/go back
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Message was not correct"))
						fmt.Println("Wrong message: " + possibleAction)
					}
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not authorized :(")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					fmt.Println("User " + strconv.FormatInt(update.Message.Chat.ID, 10) + " tried to send a message (not authorized)")
				}
			case response := <-inputChannel:
				fields := strings.Fields(response.Message)
				if fields[0] == "start" {
					msg := createMarkupForMessages(fields[1:], response.ChatId)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(response.ChatId, response.Message)
					bot.Send(msg)
				}
			}
		}
	}(updates)
	return nil
}

func turnPinOn(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action) *tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[:len(firstPart)-2]
	outputChannel <- types.Action{pin, true, chatId}
	return nil
}

func turnPinOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action) *tgbotapi.MessageConfig {
	firstPart := strings.Fields(message)[0]
	pin := firstPart[:len(firstPart)-3]
	outputChannel <- types.Action{pin, false, chatId}
	return nil
}

func removeProgrammedAction(message string, chatId int64, outputChannel chan types.ProgrammedActionOperation) *tgbotapi.MessageConfig {
	programmedAction, err := types.ProgrammedActionFromString(message, chatId)
	if err != nil {
		msg := buildMessage("Programmed action not well defined: "+err.Error(), chatId, -1)
		return &msg
	}
	outputChannel <- types.ProgrammedActionOperation{ProgrammedAction: *programmedAction, Operation: types.REMOVE}
	return nil
}

func createProgrammedAction(message string, chatId int64, outputChannel chan types.ProgrammedActionOperation) *tgbotapi.MessageConfig {
	programmedAction, err := types.ProgrammedActionFromString(message, chatId)
	if err != nil {
		msg := buildMessage("Programmed action not well defined: "+err.Error(), chatId, -1)
		return &msg
	}
	outputChannel <- types.ProgrammedActionOperation{ProgrammedAction: *programmedAction, Operation: types.CREATE}
	return nil
}

func turnPinOnAndOff(message string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action) *tgbotapi.MessageConfig {
	fields := strings.Fields(message)
	if len(fields) < 2 {
		msg := buildMessage("OnAndOff messages should contain at least two words (action and time)", chatId, replyToMessageId)
		return &msg
	}
	firstPart := fields[0]
	pin := firstPart[:len(firstPart)-8]
	duration, err := time.ParseDuration(fields[1])
	if err != nil {
		msg := buildMessage("Time not set properly", chatId, replyToMessageId)
		return &msg
	}
	outputChannel <- types.Action{pin, true, chatId}
	time.Sleep(duration)
	outputChannel <- types.Action{pin, false, chatId}
	return nil
}

type messageHandlingFunc func(action string, config configuration_loader.InitialConfiguration, chatId int64, replyToMessageId int, outputChannel chan types.Action) *tgbotapi.MessageConfig

func buildMessage(msgContent string, chatId int64, replyToMessageId int) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, msgContent)
	if replyToMessageId != -1 {
		msg.ReplyToMessageID = replyToMessageId
	}
	return msg
}

func createMarkupForMessages(messages []string, chatId int64) tgbotapi.MessageConfig {
	markup := tgbotapi.NewReplyKeyboard()
	markup.Keyboard = append(markup.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("/start")))
	for _, value := range messages {
		markup.Keyboard = append(markup.Keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(value+"On"), tgbotapi.NewKeyboardButton(value+"Off"), tgbotapi.NewKeyboardButton(value+"OnAndOff 2s")))
	}
	msg := tgbotapi.NewMessage(chatId, "Welcome to rpi bot")
	msg.ReplyMarkup = markup
	fmt.Println("User with id \"" + strconv.FormatInt(chatId, 10) + "\" requested message types")
	return msg
}
