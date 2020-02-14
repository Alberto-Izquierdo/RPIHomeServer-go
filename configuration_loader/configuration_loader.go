package configuration_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

type MyTime time.Time

type ActionTime struct {
	Action Action
	Time   MyTime
}

type Action struct {
	Pin   string
	State bool
}

type ServerConfiguration struct {
	GRPCServerPort          int
	TelegramBotToken        string
	TelegramAuthorizedUsers []int
}

type InitialConfiguration struct {
	GRPCServerIp        string
	PinsActive          []gpio_manager.PairNamePin
	ServerConfiguration *ServerConfiguration
	AutomaticMessages   []ActionTime
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

func loadConfigurationFromFileContent(fileContent []byte) (result InitialConfiguration, err error) {
	err = json.Unmarshal(fileContent, &result)
	if err == nil {
		if result.ServerConfiguration == nil && len(result.PinsActive) == 0 {
			err = errors.New("PinsActive array is empty")
		}
		for _, pin := range result.PinsActive {
			if len(strings.Fields(pin.Name)) > 1 {
				err = errors.New("Pin names should only have one word. Wrong pin: \"" + pin.Name + "\"")
			} else if matched, _ := regexp.Match("(On$)|(Off$)|(OnAndOff$)", []byte(pin.Name)); err == nil && matched {
				err = errors.New("Pin name should not end with \"On\", \"Off\" or \"OnAndOff\". Wrong pin: \"" + pin.Name + "\"")
			}
		}
		if result.ServerConfiguration != nil {
			if result.ServerConfiguration.TelegramBotToken == "" {
				err = errors.New("Telegram bot token not defined")
			}
			if len(result.ServerConfiguration.TelegramAuthorizedUsers) == 0 {
				err = errors.New("Telegram bot does not have any authorized users")
			}
			if result.ServerConfiguration.GRPCServerPort == 0 {
				err = errors.New("gRPC server port not defined")
			}
			if result.GRPCServerIp != "" {
				err = errors.New("If a node acts as server, it should not act as client (the client configuration will be done automatically")
			} else {
				result.GRPCServerIp = "localhost:" + strconv.Itoa(result.ServerConfiguration.GRPCServerPort)
			}
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
