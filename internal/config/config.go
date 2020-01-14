package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds the editor's configuration and settings.
type Config struct {
	TabSize         int
	SoftTabs        bool
	UseHighlighting bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		TabSize:         4,
		SoftTabs:        false,
		UseHighlighting: true,
	}
}

func AttoFolderPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".atto"), nil
}

// ConfigPath returns the path of Atto's config file on the current platform.
func ConfigPath() (string, error) {
	attoFolder, err := AttoFolderPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(attoFolder, "config.yml"), nil
}

// LoadConfig attempts to load the user's config
func LoadConfig() (Config, error) {
	attoFolderPath, err := AttoFolderPath()
	if err != nil {
		panic(err)
	}

	// Create the Atto folder if it doesn't exist already
	if _, err = os.Stat(attoFolderPath); os.IsNotExist(err) {
		err = os.MkdirAll(attoFolderPath, os.ModePerm)
		if err != nil {
			return DefaultConfig(), err
		}
	}

	configPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Check if the config file exists. If it does not, create one with the
	// default values.
	if _, err := os.Stat(configPath); err != nil {
		defaultConfig := DefaultConfig()

		// Marshal the default config to YAML.
		yml, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return DefaultConfig(), err
		}

		// Write the config file.
		err = ioutil.WriteFile(configPath, yml, 0644)
		if err != nil {
			return DefaultConfig(), err
		}

		return DefaultConfig(), nil
	}

	// Read the config file into memory or return the default config if there
	// is an error.
	yml, err := ioutil.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	// Unmarshal the YAML & return the default config if there is an error.
	config := Config{}
	err = yaml.Unmarshal(yml, &config)
	if err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}
