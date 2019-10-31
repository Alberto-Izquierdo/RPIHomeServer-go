package configuration_loader

import (
	"testing"
)

func TestLoadConfigurationFromPathInvalidPath(t *testing.T) {
	if _, err := LoadConfigurationFromPath("non_existing_file.json"); err == nil {
		t.Errorf("LoadConfigurationFromPath() should return an error with an invalid path")
	}
}

func TestLoadConfigurationFromPathEmptyFile(t *testing.T) {
	content := []byte("")
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() should return an error")
	}
}

func TestLoadConfigurationFromPathEmptyPinsActive(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000"
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with empty PinsActive content should return an error")
	}
}

func TestLoadConfigurationFromPathWithInvalidTypes(t *testing.T) {
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

func TestLoadConfigurationFromPathWithCorrectData(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000",
		"PinsActive": [
			{
				"name": "light",
				"pin": 	18
			}
		]
	}`)
	if config, err := loadConfigurationFromFileContent(content); err != nil {
		t.Errorf("loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	} else if config.GRPCServerIp != "192.168.2.160:8000" {
		t.Errorf("The ip was not properly loaded, it should be 192.168.2.160:8000, instead, it is %s", config.GRPCServerIp)
	} else if len(config.PinsActive) == 0 {
		t.Errorf("The array of pins should not be empty")
	} else if config.PinsActive[0].Name != "light" {
		t.Errorf("The name of the pin should be \"light\", instead it is %s", config.PinsActive[0].Name)
	} else if config.PinsActive[0].Pin != 18 {
		t.Errorf("The value of the pin should be 18, instead it is %d", config.PinsActive[0].Pin)
	}
}
