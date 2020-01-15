package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func attoFolderPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".atto"), nil
}

func configPath() (string, error) {
	attoFolder, err := attoFolderPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(attoFolder, "config.yml"), nil
}

// Config holds the editor's configuration and settings.
type Config struct {
	TabSize         int
	UseSoftTabs     bool
	UseHighlighting bool
	ShowFullPaths bool
	Use24HourTime bool
}

// Default returns the default configuration.
func Default() Config {
	return Config{
		TabSize:         4,
		UseSoftTabs:     false,
		UseHighlighting: true,
		ShowFullPaths: false,
		Use24HourTime: false,
	}
}

// Load attempts to load the user's config
func Load() (Config, error) {
	afPath, err := attoFolderPath()
	if err != nil {
		panic(err)
	}

	// Create the Atto folder if it doesn't exist already
	if _, err = os.Stat(afPath); os.IsNotExist(err) {
		err = os.MkdirAll(afPath, os.ModePerm)
		if err != nil {
			return Default(), err
		}
	}

	cfgPath, err := configPath()
	if err != nil {
		return Default(), err
	}

	// Check if the config file exists. If it does not, create one with the
	// default values.
	if _, err := os.Stat(cfgPath); err != nil {
		defaultConfig := Default()

		// Marshal the default config to YAML.
		yml, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return Default(), err
		}

		// Write the config file.
		err = ioutil.WriteFile(cfgPath, yml, 0644)
		if err != nil {
			return Default(), err
		}

		return Default(), nil
	}

	// Read the config file into memory or return the default config if there
	// is an error.
	yml, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return Default(), err
	}

	// Unmarshal the YAML & return the default config if there is an error.
	config := Config{}
	err = yaml.Unmarshal(yml, &config)
	if err != nil {
		return Default(), err
	}

	return config, nil
}
