package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Configuration holds all the needed parameters use
// the URL Shortener
type Configuration struct {
	Store    Store
	Handlers Handlers
}

// Store contains the needed fields for the Store package
type Store struct {
	DBPath          string
	ShortedIDLength uint
}

// Handlers contains the needed fields for the Handlers package
type Handlers struct {
	ListenAddr      string
	BaseURL         string
	EnableDebugMode bool
	Secret          []byte
	OAuth           struct {
		Google struct {
			ClientID     string
			ClientSecret string
		}
	}
}

var (
	config     *Configuration
	configPath string
)

// Get returns the configuration from a given file
func Get() *Configuration {
	return config
}

// Preload loads the configuration file into the memory for further usage
func Preload() error {
	var err error
	configPath, err = getConfigPath()
	if err != nil {
		return errors.Wrap(err, "could not get configuration path")
	}
	if err = updateConfig(); err != nil {
		return errors.Wrap(err, "could not update config")
	}
	return nil
}

func updateConfig() error {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.Wrap(err, "could not read configuration file")
	}
	if err = json.Unmarshal(file, &config); err != nil {
		return errors.Wrap(err, "could not unmarshal configuration file")
	}
	return nil
}

func getConfigPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "could not get executable path")
	}
	return filepath.Join(filepath.Dir(ex), "config.json"), nil
}

// Set replaces the current configuration with the given one
func Set(conf *Configuration) error {
	data, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	config = conf
	return nil
}
