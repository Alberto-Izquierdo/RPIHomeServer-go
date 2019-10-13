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
	if _, err := loadConfigurationFromFileContent(content); err != nil {
		t.Errorf("loadConfigurationFromFileContent() with proper content should not return an error, instead it returned %s", err)
	}
}
