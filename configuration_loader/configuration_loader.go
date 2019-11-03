package configuration_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

type TelegramBotConfiguration struct {
	TelegramBotToken string
	AuthorizedUsers  []int
}

type InitialConfiguration struct {
	GRPCServerIp             string
	PinsActive               []gpio_manager.PairNamePin
	TelegramBotConfiguration *TelegramBotConfiguration
}

func loadConfigurationFromFileContent(fileContent []byte) (result InitialConfiguration, err error) {
	err = json.Unmarshal(fileContent, &result)
	if err == nil {
		if len(result.PinsActive) == 0 {
			err = errors.New("PinsActive array is empty")
		}
		if result.TelegramBotConfiguration != nil {
			if len(result.TelegramBotConfiguration.AuthorizedUsers) == 0 {
				err = errors.New("Telegram bot does not have any authorized users")
			}
		}
	}
	return result, err
}

func LoadConfigurationFromPath(filePath string) (result InitialConfiguration, err error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err == nil {
		result, err = loadConfigurationFromFileContent(fileContent)
	}
	return result, err
}
