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
	if markup, ok := msg.ReplyMarkup.(tgbotapi.EditMessageReplyMarkupConfig); ok {
		if len(markup.ReplyMarkup.InlineKeyboard) != 2 {
			t.Errorf("The message should contain two rows (light and water)")
		} else if len(markup.ReplyMarkup.InlineKeyboard[0]) != 2 {
			t.Errorf("The message should contain two columns (on and off)")
		} else if markup.ReplyMarkup.InlineKeyboard[0][0].Text != "turnLightOn" {
			t.Errorf("error")
		} else if markup.ReplyMarkup.InlineKeyboard[0][1].Text != "turnLightOff" {
			t.Errorf("error")
		} else if len(markup.ReplyMarkup.InlineKeyboard[1]) != 2 {
			t.Errorf("The message should contain two columns (on and off)")
		} else if markup.ReplyMarkup.InlineKeyboard[1][0].Text != "turnWaterOn" {
			t.Errorf("error")
		} else if markup.ReplyMarkup.InlineKeyboard[1][1].Text != "turnWaterOff" {
			t.Errorf("error")
		}
	} else {
		t.Errorf("error")
	}
}

func TestTurnPinOn(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Light", 1})
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"Water", 2})
	telegramOutputChannel := make(chan configuration_loader.Action)
	go turnPinOn("turnLightOn", config, 0, 0, telegramOutputChannel)
	action := <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != true {
		t.Errorf("Action's state should be true")
	}
	go turnPinOn("turnWaterOn", config, 0, 0, telegramOutputChannel)
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
	go turnPinOff("turnLightOff", config, 0, 0, telegramOutputChannel)
	action := <-telegramOutputChannel
	if action.Pin != "Light" {
		t.Errorf("Pin name should be \"Light\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
	go turnPinOff("turnWaterOff", config, 0, 0, telegramOutputChannel)
	action = <-telegramOutputChannel
	if action.Pin != "Water" {
		t.Errorf("Pin name should be \"Water\", instead it is \"%s\"", action.Pin)
	} else if action.State != false {
		t.Errorf("Action's state should be true")
	}
}
