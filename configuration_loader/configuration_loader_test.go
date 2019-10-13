package configuration_loader

import (
	"testing"
)

func TestLoadConfigurationEmptyFile(t *testing.T) {
	content := []byte("")
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() should return an error")
	}
}

func TestLoadConfigurationEmptyPinsActive(t *testing.T) {
	content := []byte(`
	{
		"GRPCServerIp": "192.168.2.160:8000"
	}`)
	if _, err := loadConfigurationFromFileContent(content); err == nil {
		t.Errorf("loadConfigurationFromFileContent() with empty PinsActive content should return an error")
	}
}

func TestLoadConfigurationWithInvalidTypes(t *testing.T) {
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

func TestLoadConfigurationWithCorrectData(t *testing.T) {
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
	if _, err := loadConfigurationFromFileContent(content); err != nil {
		t.Errorf("loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	}
}
