package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var dataDirPath string

// ReadInConfig loads the configuration and other needed folders for further usage
func ReadInConfig() error {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("gus")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	setConfigDefaults()
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "could not reload config file")
	}
	return checkForDatadir()
}

// setConfigDefaults sets the default values for the configuration
func setConfigDefaults() {
	viper.SetDefault("http.ListenAddr", ":8080")
	viper.SetDefault("http.BaseURL", "http://localhost:3000")

	viper.SetDefault("General.DataDir", "data")
	viper.SetDefault("General.EnableDebugMode", true)
	viper.SetDefault("General.ShortedIDLength", 4)
}

// GetDataDir returns the absolute path of the data directory
func GetDataDir() string {
	return dataDirPath
}

// checkForDatadir checks for the data dir and creates it if it not exists
func checkForDatadir() error {
	var err error
	dataDirPath, err = filepath.Abs(viper.GetString("General.DataDir"))
	if err != nil {
		return errors.Wrap(err, "could not get relative data dir path")
	}
	if _, err = os.Stat(dataDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dataDirPath, 0755)
		if err != nil {
			return errors.Wrap(err, "could not create config directory")
		}
	}
	return nil
}
