package configuration_loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigurationFromPathInvalidPath(t *testing.T) {
	_, err := LoadConfigurationFromPath("non_existing_file.json")
	assert.NotEqual(t, err, nil, "LoadConfigurationFromPath() should return an error with an invalid path")
}

func TestLoadConfigurationFromStringEmptyFile(t *testing.T) {
	content := []byte("")
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() should return an error")
}

func TestLoadConfigurationFromStringEmptyPinsActive(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000"
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with empty PinsActive content should return an error")
}

func TestLoadConfigurationFromStringWithInvalidTypes(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			1234
		]
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with type mismatch should return an error")
}

func TestLoadClientConfigurationFromStringWithNoTelegramToken(t *testing.T) {
	content := []byte(`
	{
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"ServerConfiguration": {
			"TelegramAuthorizedUsers": [
				1234
			]
		}
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with not telegram token should return an error")
}

func TestLoadClientConfigurationFromStringWithNoAuthorizedUsers(t *testing.T) {
	content := []byte(`
	{
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"ServerConfiguration": {
			"TelegramBotToken": "randomToken"
		}
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with empty no authorized users should return an error")
}

func TestLoadClientConfigurationFromStringWithBothTelegramAndGrpcClientData(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"ServerConfiguration": {
			"TelegramBotToken": "randomToken",
			"TelegramAuthorizedUsers": [
				1234
			]
		}
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with both telegram bot and gRPC data should return an error")
}

func TestLoadClientConfigurationFromStringWithBothGrpcClientAndServer(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"ServerConfiguration": {
			"GRPCServerPort": 8080
		},
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		]
	}`)
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with both gRPC client and server data should return an error")
}

func TestLoadClientConfigurationFromStringWithNotGrpcServerPort(t *testing.T) {
	content := []byte(`
	{
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		],
		"ServerConfiguration": {
			"TelegramBotToken": "randomToken",
			"TelegramAuthorizedUsers": [
				1234
			]
		}
	}`)

	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with no gRPC server port should return an error")
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
		"ServerConfiguration": {
			"TelegramBotToken": "randomToken",
			"TelegramAuthorizedUsers": [
				1234
			],
			"GRPCServerPort": 8080
		}
	}`)

	config, err := loadConfigurationFromFileContent(content)
	assert.Equal(t, err, nil, "loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	assert.Equal(t, config.GRPCServerIp, "localhost:8080", "The ip should be \"localhost:8080\", instead, it is %s", config.GRPCServerIp)
	assert.NotEqual(t, len(config.PinsActive), 0, "The array of pins should not be empty")
	assert.Equal(t, config.PinsActive[0].Name, "light", "The name of the pin should be \"light\", instead it is %s", config.PinsActive[0].Name)
	assert.Equal(t, config.PinsActive[0].Pin, 18, "The value of the pin should be 18, instead it is %d", config.PinsActive[0].Pin)
	assert.NotEqual(t, config.ServerConfiguration, nil, "The server configuration should not be nil")
	assert.Equal(t, config.ServerConfiguration.TelegramBotToken, "randomToken", "The telegram bot token should be \"randomToken\", instead it is %s", config.ServerConfiguration.TelegramBotToken)
	assert.Equal(t, len(config.ServerConfiguration.TelegramAuthorizedUsers), 1, "The authorized users array should contain one elements")
	assert.Equal(t, config.ServerConfiguration.TelegramAuthorizedUsers[0], 1234, "The authorized users array should contain 1234, instead it contains %d", config.ServerConfiguration.TelegramAuthorizedUsers[0])
	assert.Equal(t, config.ServerConfiguration.GRPCServerPort, 8080, "The gRPC server por should be 8080, instead it is %d", config.ServerConfiguration.GRPCServerPort)
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
	_, err := loadConfigurationFromFileContent(content)
	assert.NotEqual(t, err, nil, "loadConfigurationFromFileContent() with automatic messages to pins not present in the gpio manager should return an error")
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

	config, err := loadConfigurationFromFileContent(content)
	assert.Equal(t, err, nil, "loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	assert.Equal(t, config.GRPCServerIp, "192.168.2.160:8000", "The ip should be empty, instead, it is %s", config.GRPCServerIp)
	assert.NotEqual(t, len(config.PinsActive), 0, "The array of pins should not be empty")
	assert.Equal(t, config.PinsActive[0].Name, "light", "The name of the pin should be \"light\", instead it is %s", config.PinsActive[0].Name)
	assert.Equal(t, config.PinsActive[0].Pin, 18, "The value of the pin should be 18, instead it is %d", config.PinsActive[0].Pin)
	assert.NotEqual(t, config.ServerConfiguration, nil, "The telegram bot configuration should be nil")
}

func TestLoadClientConfigurationFromStringWithComplexPinName(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "bedroom light",
				"pin": 	18
			}
		]
	}`)

	_, err := loadConfigurationFromFileContent(content)
	assert.NotNil(t, err, "loadConfigurationFromFileContent() with complex pin names should return an error")
}
