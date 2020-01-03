package telegram_bot

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
)

func TestLaunchTelegramBot(t *testing.T) {
	telegramInputChannel := make(chan string)
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
	LaunchTelegramBot(config, telegramOutputChannel, telegramInputChannel, telegramExitChannel)
}
