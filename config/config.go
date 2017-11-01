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
	ShortedIDLength int
}

// Handlers contains the needed fields for the Handlers package
type Handlers struct {
	ListenAddr         string
	EnableGinDebugMode bool
	OAuth              struct {
		Google struct {
			ClientID     string
			ClientSecret string
		}
	}
}

// Get returns the configuration from a given file
func Get() (*Configuration, error) {
	var config *Configuration
	ex, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "could not get executable path")
	}
	file, err := ioutil.ReadFile(filepath.Join(filepath.Dir(ex), "config.json"))
	if err != nil {
		return nil, errors.Wrap(err, "could not read configuration file")
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal configuration file")
	}
	return config, nil
}
