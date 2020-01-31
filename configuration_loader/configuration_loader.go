package configuration_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

type TelegramBotConfiguration struct {
	TelegramBotToken string
	AuthorizedUsers  []int
}

type MyTime time.Time

type ActionTime struct {
	Action Action
	Time   MyTime
}

type Action struct {
	Pin   string
	State bool
}

type GRPCServerConfiguration struct {
	Port int
}

type InitialConfiguration struct {
	GRPCServerIp             string
	PinsActive               []gpio_manager.PairNamePin
	TelegramBotConfiguration *TelegramBotConfiguration
	GRPCServerConfiguration  *GRPCServerConfiguration
	AutomaticMessages        []ActionTime
}

func (a *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	*a = MyTime(t)
	return nil
}

func (a MyTime) Format(s string) string {
	t := time.Time(a)
	return t.Format(s)
}

func (this ActionTime) LessThan(other interface{}) bool {
	return time.Time(this.Time).Before(time.Time(other.(ActionTime).Time))
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
		} else if result.GRPCServerIp != "" && result.GRPCServerConfiguration != nil {
			err = errors.New("Configuration should only contain one of gRPC client or gRPC server")
		}
		if len(result.AutomaticMessages) > 0 {
			for index, automaticMessage := range result.AutomaticMessages {
				found := false
				for _, pin := range result.PinsActive {
					if pin.Name == automaticMessage.Action.Pin {
						found = true
						break
					}
				}
				if !found {
					err = errors.New("Automatic message number " + strconv.Itoa(index) + ", " + automaticMessage.Action.Pin + " not present in the pins active")
				}
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
