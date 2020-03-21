package telegram_bot

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

func TestWrongConfig(t *testing.T) {
	telegramOutputChannel := make(chan types.Action)
	telegramInputChannel := make(chan types.TelegramMessage)
	telegramExitChannel := make(chan bool)
	var config configuration_loader.InitialConfiguration
	var serverConfig configuration_loader.ServerConfiguration
	config.ServerConfiguration = &serverConfig
	config.ServerConfiguration.TelegramBotToken = "asdf"
	config.ServerConfiguration.TelegramAuthorizedUsers = append(config.ServerConfiguration.TelegramAuthorizedUsers, 1234, 5678)
	err := LaunchTelegramBot(config, telegramOutputChannel, nil, telegramInputChannel, telegramExitChannel)
	assert.NotEqual(t, err, nil, "Wrong config should return an error")
}

func TestLaunchTelegramBot(t *testing.T) {
	telegramOutputChannel := make(chan types.Action)
	telegramInputChannel := make(chan types.TelegramMessage)
	telegramExitChannel := make(chan bool)
	var config configuration_loader.InitialConfiguration
	var serverConfig configuration_loader.ServerConfiguration
	config.ServerConfiguration = &serverConfig
	config.ServerConfiguration.TelegramBotToken = "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I"
	config.ServerConfiguration.TelegramAuthorizedUsers = append(config.ServerConfiguration.TelegramAuthorizedUsers, 1234, 5678)
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Light", 1})
	go func() {
		time.Sleep(2 * time.Second)
		telegramExitChannel <- true
		<-telegramExitChannel
		close(telegramExitChannel)
	}()
	LaunchTelegramBot(config, telegramOutputChannel, nil, telegramInputChannel, telegramExitChannel)
}

func TestGetMessagesAvailableMarkup(t *testing.T) {
	messages := []string{"Light", "Water"}
	msg := createMarkupForMessages(messages, 0)
	markup, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup)
	assert.True(t, ok, "Error getting the message's reply markup")
	assert.Equal(t, len(markup.Keyboard), 3, "The message should contain three rows (/start, light and water)")
	assert.Equal(t, len(markup.Keyboard[0]), 1, "The message should contain one column (/start)")
	assert.Equal(t, markup.Keyboard[0][0].Text, "/start", "Button should contain \"/start\" and it is \"%s\"", markup.Keyboard[0][0].Text)
	assert.Equal(t, len(markup.Keyboard[1]), 3, "The message should contain three columns (on, off, onAndOff)")
	assert.Equal(t, markup.Keyboard[1][0].Text, "LightOn", "Button should contain \"LightOn\" and it is \"%s\"", markup.Keyboard[1][0].Text)
	assert.Equal(t, markup.Keyboard[1][1].Text, "LightOff", "Button should contain \"LightOff\" and it is \"%s\"", markup.Keyboard[1][1].Text)
	assert.Equal(t, markup.Keyboard[1][2].Text, "LightOnAndOff 2s", "Button should contain \"LightOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[1][2].Text)
	assert.Equal(t, len(markup.Keyboard[2]), 3, "The message should contain three columns (on, off, onAndOff)")
	assert.Equal(t, markup.Keyboard[2][0].Text, "WaterOn", "Button should contain \"WaterOn\" and it is \"%s\"", markup.Keyboard[2][0].Text)
	assert.Equal(t, markup.Keyboard[2][1].Text, "WaterOff", "Button should contain \"WaterOff\" and it is \"%s\"", markup.Keyboard[2][1].Text)
	assert.Equal(t, markup.Keyboard[2][2].Text, "WaterOnAndOff 2s", "Button should contain \"WaterOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[2][2].Text)
}

func TestTurnPinOn(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan types.Action)
	go func() {
		turnPinOn("LightOn", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Light", "Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, true, "Action's state should be true")
	go func() {
		turnPinOn("WaterOn", config, 0, 0, telegramOutputChannel)
	}()
	action = <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Water", "Pin name should be \"Water\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, true, "Action's state should be true")
}

func TestTurnPinOff(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan types.Action)
	go func() {
		turnPinOff("LightOff", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Light", "Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, false, "Action's state should be true")
	go func() {
		turnPinOff("WaterOff", config, 0, 0, telegramOutputChannel)
	}()
	action = <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Water", "Pin name should be \"Water\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, false, "Action's state should be true")
}

func TestTurnPinOnAndOff(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, types.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan types.Action)
	msg := turnPinOnAndOff("LightOnAndOff", config, 0, 0, telegramOutputChannel)
	assert.Equal(t, msg.Text, "OnAndOff messages should contain at least two words (action and time)", "Wrong message should return an error")
	msg = turnPinOnAndOff("LightOnAndOff 40w", config, 0, 0, telegramOutputChannel)
	assert.Equal(t, msg.Text, "Time not set properly", "Wrong time format should return an error")
	go func() {
		turnPinOnAndOff("LightOnAndOff 1s", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Light", "Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, true, "Action's state should be false")
	action = <-telegramOutputChannel
	assert.Equal(t, action.Pin, "Light", "Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	assert.Equal(t, action.State, false, "Action's state should be true")
}
