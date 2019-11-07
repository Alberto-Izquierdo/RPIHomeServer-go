package configuration_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

type TelegramBotConfiguration struct {
	TelegramBotToken string
	AuthorizedUsers  []int
}

type ActionTime time.Time

type Action struct {
	Pin   string
	State bool
	Time  ActionTime
}

type InitialConfiguration struct {
	GRPCServerIp             string
	PinsActive               []gpio_manager.PairNamePin
	TelegramBotConfiguration *TelegramBotConfiguration
	AutomaticMessages        []Action
}

func (a *ActionTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	*a = ActionTime(t)
	return nil
}

func (a ActionTime) Format(s string) string {
	t := time.Time(a)
	return t.Format(s)
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
			if result.GRPCServerIp != "" {
				err = errors.New("Configuration should only contain one of telegram bot configuration or gRPC configuration")
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
