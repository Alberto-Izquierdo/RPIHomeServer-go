package telegram_bot

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestWrongConfig(t *testing.T) {
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramInputChannel := make(chan string)
	telegramExitChannel := make(chan bool)
	var config configuration_loader.InitialConfiguration
	var serverConfig configuration_loader.ServerConfiguration
	config.ServerConfiguration = &serverConfig
	config.ServerConfiguration.TelegramBotToken = "asdf"
	config.ServerConfiguration.TelegramAuthorizedUsers = append(config.ServerConfiguration.TelegramAuthorizedUsers, 1234, 5678)
	err := LaunchTelegramBot(config, telegramOutputChannel, telegramInputChannel, telegramExitChannel)
	if err == nil {
		t.Errorf("Wrong config should return an error")
	}
}

func TestLaunchTelegramBot(t *testing.T) {
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramInputChannel := make(chan string)
	telegramExitChannel := make(chan bool)
	var config configuration_loader.InitialConfiguration
	var serverConfig configuration_loader.ServerConfiguration
	config.ServerConfiguration = &serverConfig
	config.ServerConfiguration.TelegramBotToken = "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I"
	config.ServerConfiguration.TelegramAuthorizedUsers = append(config.ServerConfiguration.TelegramAuthorizedUsers, 1234, 5678)
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	go func() {
		time.Sleep(2 * time.Second)
		telegramExitChannel <- true
		<-telegramExitChannel
		close(telegramExitChannel)
	}()
	LaunchTelegramBot(config, telegramOutputChannel, telegramInputChannel, telegramExitChannel)
}

func TestGetMessagesAvailableMarkup(t *testing.T) {
	messages := []string{"Light", "Water"}
	msg := createMarkupForMessages(messages, 0, 0)
	if markup, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup); ok {
		if len(markup.Keyboard) != 3 {
			t.Errorf("The message should contain three rows (/start, light and water)")
		} else if len(markup.Keyboard[0]) != 1 {
			t.Errorf("The message should contain one column (/start)")
		} else if markup.Keyboard[0][0].Text != "/start" {
			t.Errorf("Button should contain \"/start\" and it is \"%s\"", markup.Keyboard[0][0].Text)
		} else if len(markup.Keyboard[1]) != 3 {
			t.Errorf("The message should contain three columns (on, off, onAndOff)")
		} else if markup.Keyboard[1][0].Text != "LightOn" {
			t.Errorf("Button should contain \"LightOn\" and it is \"%s\"", markup.Keyboard[1][0].Text)
		} else if markup.Keyboard[1][1].Text != "LightOff" {
			t.Errorf("Button should contain \"LightOff\" and it is \"%s\"", markup.Keyboard[1][1].Text)
		} else if markup.Keyboard[1][2].Text != "LightOnAndOff 2s" {
			t.Errorf("Button should contain \"LightOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[1][2].Text)
		} else if len(markup.Keyboard[2]) != 3 {
			t.Errorf("The message should contain three columns (on, off, onAndOff)")
		} else if markup.Keyboard[2][0].Text != "WaterOn" {
			t.Errorf("Button should contain \"WaterOn\" and it is \"%s\"", markup.Keyboard[2][0].Text)
		} else if markup.Keyboard[2][1].Text != "WaterOff" {
			t.Errorf("Button should contain \"WaterOff\" and it is \"%s\"", markup.Keyboard[2][1].Text)
		} else if markup.Keyboard[2][2].Text != "WaterOnAndOff 2s" {
			t.Errorf("Button should contain \"WaterOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[2][2].Text)
		}
	} else {
		t.Errorf("Error getting the message's reply markup")
	}
}

func TestTurnPinOn(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramInputChannel := make(chan string)
	go func() {
		turnPinOn("LightOn", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	}()
	action := <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be true")
	}
	go func() {
		turnPinOn("WaterOn", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	}()
	action = <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Water" {
		t.Errorf("Pin name should be \"Water\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be true")
	}
}

func TestTurnPinOff(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramInputChannel := make(chan string)
	go func() {
		turnPinOff("LightOff", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	}()
	action := <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
	go func() {
		turnPinOff("WaterOff", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	}()
	action = <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Water" {
		t.Errorf("Pin name should be \"Water\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
}

func TestTurnPinOnAndOff(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramInputChannel := make(chan string)
	msg := turnPinOnAndOff("LightOnAndOff", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	if msg.Text != "OnAndOff messages should contain at least two words (action and time)" {
		t.Errorf("Wrong message should return an error")
	}
	msg = turnPinOnAndOff("LightOnAndOff 40w", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	if msg.Text != "Time not set properly" {
		t.Errorf("Wrong time format should return an error")
	}
	go func() {
		turnPinOnAndOff("LightOnAndOff 1s", config, 0, 0, telegramOutputChannel, telegramInputChannel)
	}()
	action := <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be false")
	}
	action = <-telegramOutputChannel
	telegramInputChannel <- "test"
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
}
