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
	Schema   string   `json:"$schema"`
	Store    Store    `description:"Store holds the configuration values for the storage package"`
	Handlers Handlers `description:"Handlers holds the configuration for the handlers package"`
}

// Store contains the needed fields for the Store package
type Store struct {
	DBPath          string `description:"relative or absolute path of your bolt DB"`
	ShortedIDLength uint   `description:"Length of the random generated ID which is used for new shortened URLs"`
}

// Handlers contains the needed fields for the Handlers package
type Handlers struct {
	ListenAddr      string `description:"Consists of 'IP:Port', normally the value ':8080' e.g. is enough"`
	BaseURL         string `description:"Required for the authentification via OAuth. E.g. 'http://mydomain.com'"`
	EnableDebugMode bool   `description:"Activates more detailed logging to the stdout"`
	Secret          []byte `description:"Used for encryption of the JWT and for the CookieJar. Will be randomly generated when it isn't set"`
	OAuth           struct {
		Google struct {
			ClientID     string `description:"ClientID which you get from console.cloud.google.com"`
			ClientSecret string `description:"ClientSecret which get from console.cloud.google.com"`
		} `description:"Google holds the OAuth configuration for the Google provider"`
	} `description:"OAuth holds the OAuth specific settings"`
}

var config *Configuration

var configPath string

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
