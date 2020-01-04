package telegram_bot

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestLaunchTelegramBot(t *testing.T) {
	telegramOutputChannel := make(chan configuration_loader.Action)
	telegramExitChannel := make(chan bool)
	var config configuration_loader.InitialConfiguration
	var telegramConfig configuration_loader.TelegramBotConfiguration
	config.TelegramBotConfiguration = &telegramConfig
	config.TelegramBotConfiguration.TelegramBotToken = "153667468:AAHlSHlMqSt1f_uFmVRJbm5gntu2HI4WW8I"
	config.TelegramBotConfiguration.AuthorizedUsers = append(config.TelegramBotConfiguration.AuthorizedUsers, 1234, 5678)
	go func() {
		time.Sleep(2 * time.Second)
		close(telegramExitChannel)
	}()
	LaunchTelegramBot(config, telegramOutputChannel, telegramExitChannel)
}

func TestGetMessagesAvailableMarkup(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Water", 2})
	msg := getMessagesAvailableMarkup("", config, 0, 0, nil)
	if markup, ok := msg.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup); ok {
		if len(markup.Keyboard) != 2 {
			t.Errorf("The message should contain two rows (light and water)")
		} else if len(markup.Keyboard[0]) != 3 {
			t.Errorf("The message should contain three columns (on, off, onAndOff)")
		} else if markup.Keyboard[0][0].Text != "turnLightOn" {
			t.Errorf("Button should contain \"turnLightOn\" and it is \"%s\"", markup.Keyboard[0][0].Text)
		} else if markup.Keyboard[0][1].Text != "turnLightOff" {
			t.Errorf("Button should contain \"turnLightOff\" and it is \"%s\"", markup.Keyboard[0][1].Text)
		} else if markup.Keyboard[0][2].Text != "turnLightOnAndOff 2s" {
			t.Errorf("Button should contain \"turnLightOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[0][2].Text)
		} else if len(markup.Keyboard[1]) != 3 {
			t.Errorf("The message should contain three columns (on, off, onAndOff)")
		} else if markup.Keyboard[1][0].Text != "turnWaterOn" {
			t.Errorf("Button should contain \"turnWaterOn\" and it is \"%s\"", markup.Keyboard[1][0].Text)
		} else if markup.Keyboard[1][1].Text != "turnWaterOff" {
			t.Errorf("Button should contain \"turnWaterOff\" and it is \"%s\"", markup.Keyboard[1][1].Text)
		} else if markup.Keyboard[1][2].Text != "turnWaterOnAndOff 2s" {
			t.Errorf("Button should contain \"turnWaterOnAndOff 2s\" and it is \"%s\"", markup.Keyboard[1][2].Text)
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
	go func() {
		turnPinOn("turnLightOn", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be true")
	}
	go func() {
		turnPinOn("turnWaterOn", config, 0, 0, telegramOutputChannel)
	}()
	action = <-telegramOutputChannel
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
	go func() {
		turnPinOff("turnLightOff", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
	go func() {
		turnPinOff("turnWaterOff", config, 0, 0, telegramOutputChannel)
	}()
	action = <-telegramOutputChannel
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
	go func() {
		turnPinOnAndOff("turnLightOnAndOff 1s", config, 0, 0, telegramOutputChannel)
	}()
	action := <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be false")
	}
	action = <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
}
