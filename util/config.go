package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	dataDirPath string
	// DoNotSetConfigName is used to predefine if the name of the config should be set.
	// Used for unit testing
	DoNotSetConfigName = false
)

// ReadInConfig loads the configuration and other needed folders for further usage
func ReadInConfig() error {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("gus")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if !DoNotSetConfigName {
		viper.SetConfigName("config")
	}
	viper.AddConfigPath(".")
	setConfigDefaults()
	switch err := viper.ReadInConfig(); err.(type) {
	case viper.ConfigFileNotFoundError:
		logrus.Info("No configuration file found, using defaults and environment overrides.")
		break
	case nil:
		break
	default:
		return errors.Wrap(err, "could not read config file")
	}
	return checkForDatadir()
}

// setConfigDefaults sets the default values for the configuration
func setConfigDefaults() {
	viper.SetDefault("listen_addr", ":8080")
	viper.SetDefault("base_url", "http://localhost:3000")

	viper.SetDefault("data_dir", "data")
	viper.SetDefault("enable_debug_mode", true)
	viper.SetDefault("shorted_id_length", 4)
}

// GetDataDir returns the absolute path of the data directory
func GetDataDir() string {
	return dataDirPath
}

// checkForDatadir checks for the data dir and creates it if it not exists
func checkForDatadir() error {
	var err error
	dataDirPath, err = filepath.Abs(viper.GetString("data_dir"))
	if err != nil {
		return errors.Wrap(err, "could not get relative data dir path")
	}
	if _, err = os.Stat(dataDirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dataDirPath, 0755); err != nil {
			return errors.Wrap(err, "could not create config directory")
		}
	}
	return nil
}
