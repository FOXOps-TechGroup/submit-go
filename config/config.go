package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FOXOps-TechGroup/submit-go/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Submitter string `yaml:"submitter"`
	// Online Judge system's URL,with port like:
	//http://127.0.0.1:4567
	URL string `yaml:"url"`
	//if system require password
	UserID   string `yaml:"user_id"`
	Password string `yaml:"password"`
}

var GlobalConfig = &Config{}

type EditSettingsError struct {
	configPath string
}

func (e EditSettingsError) Error() string {
	return fmt.Sprintf("created example configuration file at %s, please edit it with your settings",
		e.configPath)
}

// Read loads configuration from $HOME/.config/submit/config.yaml
// If the file doesn't exist, it creates a example one and returns an error
func Read() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "submit")
	configPath := filepath.Join(configDir, "config.yaml")

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		defaultConfig := Config{
			Submitter: "",
			URL:       "",
			UserID:    "",
			Password:  "",
		}

		yamlData, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to create example config: %w", err)
		}

		if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		return EditSettingsError{configPath: configDir}
	} else if err != nil {
		return fmt.Errorf(
			"failed to check if config file exists: %w",
			err)
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, GlobalConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	utils.Logger.Debug("config file read", "config", GlobalConfig)

	return nil
}
