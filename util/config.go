package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	envstruct "github.com/mxschmitt/golang-env-struct"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configuration are the available config values
type Configuration struct {
	ListenAddr      string        `yaml:"ListenAddr" env:"LISTEN_ADDR"`
	BaseURL         string        `yaml:"BaseURL" env:"BASE_URL"`
	DataDir         string        `yaml:"DataDir" env:"DATA_DIR"`
	Backend         string        `yaml:"Backend" env:"BACKEND"`
	RedisHost       string        `yaml:"RedisHost" env:"REDIS_HOST"`
	RedisPassword   string        `yaml:"RedisPassword" env:"REDIS_PASSWORD"`
	AuthBackend     string        `yaml:"AuthBackend" env:"AUTH_BACKEND"`
	UseSSL          bool          `yaml:"EnableSSL" env:"USE_SSL"`
	EnableDebugMode bool          `yaml:"EnableDebugMode" env:"ENABLE_DEBUG_MODE"`
	EnableColorLogs bool          `yaml:"EnableColorLogs" env:"ENABLE_COLOR_LOGS"`
	ShortedIDLength int           `yaml:"ShortedIDLength" env:"SHORTED_ID_LENGTH"`
	Google          oAuthConf     `yaml:"Google" env:"GOOGLE"`
	GitHub          oAuthConf     `yaml:"GitHub" env:"GITHUB"`
	Microsoft       oAuthConf     `yaml:"Microsoft" env:"MICROSOFT"`
	Proxy           proxyAuthConf `yaml:"Proxy" env:"PROXY"`
}

type oAuthConf struct {
	ClientID     string `yaml:"ClientID" env:"CLIENT_ID"`
	ClientSecret string `yaml:"ClientSecret" env:"CLIENT_SECRET"`
}

type proxyAuthConf struct {
	RequireUserHeader bool   `yaml:"RequireUserHeader" env:"REQUIRE_USER_HEADER"`
	UserHeader        string `yaml:"UserHeader" env:"USER_HEADER"`
	DisplayNameHeader string `yaml:"DisplayNameHeader" env:"DISPLAY_NAME_HEADER"`
}

// config contains the default values
var Config = Configuration{
	ListenAddr:      ":8080",
	BaseURL:         "http://localhost:3000",
	DataDir:         "data",
	Backend:         "boltdb",
	EnableDebugMode: false,
	EnableColorLogs: true,
	UseSSL:          false,
	ShortedIDLength: 4,
	AuthBackend:     "oauth",
}

// ReadInConfig loads the Configuration and other needed folders for further usage
func ReadInConfig() error {
	file, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(file, &Config); err != nil {
			return errors.Wrap(err, "could not unmarshal yaml file")
		}
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "could not read config file")
	} else {
		logrus.Info("No configuration file found, using defaults with environment variable overrides.")
	}
	if err := envstruct.ApplyEnvVars(&Config, "GUS"); err != nil {
		return errors.Wrap(err, "could not apply environment configuration")
	}
	logrus.Info("Loaded configuration: %v", Config)
	Config.DataDir, err = filepath.Abs(Config.DataDir)
	if err != nil {
		return errors.Wrap(err, "could not get relative data dir path")
	}
	if _, err = os.Stat(Config.DataDir); os.IsNotExist(err) {
		if err = os.MkdirAll(Config.DataDir, 0755); err != nil {
			return errors.Wrap(err, "could not create config directory")
		}
	}
	return nil
}

func (o oAuthConf) Enabled() bool {
	return o.ClientSecret != ""
}

// GetConfig returns the configuration from the memory
func GetConfig() Configuration {
	return Config
}

// SetConfig sets the configuration into the memory
func SetConfig(c Configuration) {
	Config = c
}
