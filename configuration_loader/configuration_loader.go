package configuration_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

type InitialConfiguration struct {
	gRPCServerIp string
	PinsActive   []gpio_manager.PairNamePin
}

func loadConfigurationFromFileContent(fileContent []byte) (result InitialConfiguration, err error) {
	err = json.Unmarshal(fileContent, &result)
	if err == nil {
		if len(result.PinsActive) == 0 {
			err = errors.New("PinsActive array is empty")
		}
	}
	return result, err
}

func LoadConfigurationFromPath(filePath string) (result InitialConfiguration, err error) {
	if fileContent, err := ioutil.ReadFile(filePath); err == nil {
		result, err = loadConfigurationFromFileContent(fileContent)
	}
	return result, err
}
