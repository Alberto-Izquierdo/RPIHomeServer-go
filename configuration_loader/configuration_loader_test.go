package configuration_loader

import (
	"testing"
)

func TestLoadConfigurationFromPathInvalidPath(t *testing.T) {
	if _, err := LoadConfigurationFromPath("non_existing_file.json"); err == nil {
		t.Errorf("LoadConfigurationFromPath() should return an error with an invalid path")
	}
}

func TestLoadConfigurationFromStringEmptyFile(t *testing.T) {
	content := []byte("")
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() should return an error")
	}
}

func TestLoadConfigurationFromStringEmptyPinsActive(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000"
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with empty PinsActive content should return an error")
	}
}

func TestLoadConfigurationFromStringWithInvalidTypes(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			1234
		]
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with type mismatch should return an error")
	}
}

func TestLoadClientConfigurationFromStringWithNoAuthorizedUsers(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"TelegramBotConfiguration": {
			"TelegramBotToken": "randomToken"
		}
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with empty no authorized users should return an error")
	}
}

func TestLoadClientConfigurationFromStringWithBothTelegramAndGRPCData(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"TelegramBotConfiguration": {
			"TelegramBotToken": "randomToken",
			"AuthorizedUsers": [
				1234
			]
		}
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with both telegram bot and gRPC data should return an error, instead it returned %s", err)
	}
}

func TestLoadClientConfigurationFromStringWithCorrectTelegramBotData(t *testing.T) {
	content := []byte(`
	{
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"TelegramBotConfiguration": {
			"TelegramBotToken": "randomToken",
			"AuthorizedUsers": [
				1234
			]
		}
	}`)

	if config, err := loadConfigurationFromFileContent(content); err != nil {
		t.Errorf("loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	} else if config.GRPCServerIp != "" {
		t.Errorf("The ip should be empty, instead, it is %s", config.GRPCServerIp)
	} else if len(config.PinsActive) == 0 {
		t.Errorf("The array of pins should not be empty")
	} else if config.PinsActive[0].Name != "light" {
		t.Errorf("The name of the pin should be \"light\", instead it is %s", config.PinsActive[0].Name)
	} else if config.PinsActive[0].Pin != 18 {
		t.Errorf("The value of the pin should be 18, instead it is %d", config.PinsActive[0].Pin)
	} else if config.TelegramBotConfiguration == nil {
		t.Errorf("The telegram bot configuration should not be nil")
	} else if config.TelegramBotConfiguration.TelegramBotToken != "randomToken" {
		t.Errorf("The telegram bot token should be \"randomToken\", instead it is %s", config.TelegramBotConfiguration.TelegramBotToken)
	} else if len(config.TelegramBotConfiguration.AuthorizedUsers) != 1 {
		t.Errorf("The authorized users array should contain one elements")
	} else if config.TelegramBotConfiguration.AuthorizedUsers[0] != 1234 {
		t.Errorf("The authorized users array should contain 1234, instead it contains %d", config.TelegramBotConfiguration.AuthorizedUsers[0])
	}

}

func TestLoadClientConfigurationFromStringWithWrongAutomaticMessages(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"AutomaticMessages": [
			{
				"Action": {
					"Pin": "water",
					"State": true
				},
				"Time": "03:45:10"
			},
			{
				"Action": {
					"Pin": "water",
					"State": false
				},
				"Time": "03:45:15"
			}
		]
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with automatic messages to pins not present in the gpio manager should return an error")
	}
}

func TestLoadClientConfigurationFromStringWithCorrectGRPCData(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"AutomaticMessages": [
			{
				"Action": {
					"Pin": "light",
					"State": true
				},
				"Time": "03:45:10"
			},
			{
				"Action": {
					"Pin": "light",
					"State": false
				},
				"Time": "03:45:15"
			}
		]
	}`)

	if config, err := loadConfigurationFromFileContent(content); err != nil {
		t.Errorf("loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	} else if config.GRPCServerIp != "192.168.2.160:8000" {
		t.Errorf("The ip should be empty, instead, it is %s", config.GRPCServerIp)
	} else if len(config.PinsActive) == 0 {
		t.Errorf("The array of pins should not be empty")
	} else if config.PinsActive[0].Name != "light" {
		t.Errorf("The name of the pin should be \"light\", instead it is %s", config.PinsActive[0].Name)
	} else if config.PinsActive[0].Pin != 18 {
		t.Errorf("The value of the pin should be 18, instead it is %d", config.PinsActive[0].Pin)
	} else if config.TelegramBotConfiguration != nil {
		t.Errorf("The telegram bot configuration should be nil")
	}
}
