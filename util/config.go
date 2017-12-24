package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	ListenAddr      string    `yaml:"ListenAddr" env:"LISTEN_ADDR"`
	BaseURL         string    `yaml:"BaseURL" env:"BASE_URL"`
	DataDir         string    `yaml:"DataDir" env:"DATA_DIR"`
	UseSSL          bool      `yaml:"EnableSSL" env:"USE_SSL"`
	EnableDebugMode bool      `yaml:"EnableDebugMode" env:"ENABLE_DEBUG_MODE"`
	ShortedIDLength int       `yaml:"ShortedIDLength" env:"SHORTED_ID_LENGTH"`
	Google          oAuthConf `yaml:"Google" env:"GOOGLE"`
	GitHub          oAuthConf `yaml:"GitHub" env:"GITHUB"`
	Microsoft       oAuthConf `yaml:"Microsoft" env:"MICROSOFT"`
}

type oAuthConf struct {
	ClientID     string `yaml:"ClientID" env:"CLIENT_ID"`
	ClientSecret string `yaml:"ClientSecret" env:"CLIENT_SECRET"`
}

var (
	config = Configuration{
		ListenAddr:      ":8080",
		BaseURL:         "http://localhost:3000",
		DataDir:         "data",
		EnableDebugMode: false,
		UseSSL:          false,
		ShortedIDLength: 4,
	}
)

// ReadInConfig loads the Configuration and other needed folders for further usage
func ReadInConfig() error {
	file, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(file, &config); err != nil {
			return errors.Wrap(err, "could not unmarshal yaml file")
		}
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "could not read config file")
	} else {
		logrus.Info("No configuration file found, using defaults with environment variable overrides.")
	}
	if err := config.ApplyEnvironmentConfig(); err != nil {
		return errors.Wrap(err, "could not apply environment configuration")
	}
	config.DataDir, err = filepath.Abs(config.DataDir)
	if err != nil {
		return errors.Wrap(err, "could not get relative data dir path")
	}
	if _, err = os.Stat(config.DataDir); os.IsNotExist(err) {
		if err = os.MkdirAll(config.DataDir, 0755); err != nil {
			return errors.Wrap(err, "could not create config directory")
		}
	}
	return nil
}

func (c *Configuration) ApplyEnvironmentConfig() error {
	return c.setDefaultValue(reflect.ValueOf(c), reflect.TypeOf(*c), -1, "GUS")
}

func (c *Configuration) setDefaultValue(v reflect.Value, t reflect.Type, counter int, prefix string) error {
	if v.Kind() != reflect.Ptr {
		return errors.New("Not a pointer value")
	}
	f := reflect.StructField{}
	if counter != -1 {
		f = t.Field(counter)
	}
	v = reflect.Indirect(v)
	fieldEnv, exists := f.Tag.Lookup("env")
	env := os.Getenv(prefix + fieldEnv)
	if exists && env != "" {
		switch v.Kind() {
		case reflect.Int:
			envI, err := strconv.Atoi(env)
			if err != nil {
				logrus.Warningf("could not parse to int: %v", err)
				break
			}
			v.SetInt(int64(envI))
		case reflect.String:
			v.SetString(env)
		case reflect.Bool:
			envB, err := strconv.ParseBool(env)
			if err != nil {
				logrus.Warningf("could not parse to bool: %v", err)
				break
			}
			v.SetBool(envB)
		}
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if err := c.setDefaultValue(v.Field(i).Addr(), v.Type(), i, prefix+fieldEnv+"_"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o oAuthConf) Enabled() bool {
	return o.ClientSecret != ""
}

func GetConfig() Configuration {
	return config
}

func SetConfig(c Configuration) {
	config = c
}
