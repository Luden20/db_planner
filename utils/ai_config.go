package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const defaultOpenAIModel = "gpt-5-mini"

type AppConfig struct {
	OpenAIAPIKey string `json:"openAIApiKey"`
	OpenAIModel  string `json:"openAIModel"`
}

type AISettings struct {
	HasAPIKey   bool
	OpenAIModel string
}

type SQLGenerationResult struct {
	Database   string
	Model      string
	SQL        string
	ExportJSON string
}

func DefaultOpenAIModel() string {
	return defaultOpenAIModel
}

func LoadAppConfig() (*AppConfig, error) {
	configPath, err := appConfigPath()
	if err != nil {
		return nil, err
	}

	config := &AppConfig{
		OpenAIModel: defaultOpenAIModel,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return config, nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	config.normalize()
	return config, nil
}

func SaveOpenAIAPIKey(apiKey string) (*AISettings, error) {
	config, err := LoadAppConfig()
	if err != nil {
		return nil, err
	}

	config.OpenAIAPIKey = strings.TrimSpace(apiKey)
	config.normalize()

	if err := saveAppConfig(config); err != nil {
		return nil, err
	}

	return config.ToAISettings(), nil
}

func GetAISettings() (*AISettings, error) {
	config, err := LoadAppConfig()
	if err != nil {
		return nil, err
	}
	return config.ToAISettings(), nil
}

func (c *AppConfig) ToAISettings() *AISettings {
	if c == nil {
		return &AISettings{
			HasAPIKey:   false,
			OpenAIModel: defaultOpenAIModel,
		}
	}

	c.normalize()
	return &AISettings{
		HasAPIKey:   strings.TrimSpace(c.OpenAIAPIKey) != "",
		OpenAIModel: c.OpenAIModel,
	}
}

func (c *AppConfig) normalize() {
	if c.OpenAIModel == "" {
		c.OpenAIModel = defaultOpenAIModel
	}
}

func saveAppConfig(config *AppConfig) error {
	configPath, err := appConfigPath()
	if err != nil {
		return err
	}

	config.normalize()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func appConfigPath() (string, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configRoot, "db_planner")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "settings.json"), nil
}
