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
	ListenAddr       string        `yaml:"ListenAddr" env:"LISTEN_ADDR"`
	BaseURL          string        `yaml:"BaseURL" env:"BASE_URL"`
	DisplayURL       string        `yaml:"DisplayURL" env:"DISPLAY_URL"`
	DataDir          string        `yaml:"DataDir" env:"DATA_DIR"`
	Backend          string        `yaml:"Backend" env:"BACKEND"`
	AuthBackend      string        `yaml:"AuthBackend" env:"AUTH_BACKEND"`
	UseSSL           bool          `yaml:"EnableSSL" env:"USE_SSL"`
	EnableDebugMode  bool          `yaml:"EnableDebugMode" env:"ENABLE_DEBUG_MODE"`
	EnableAccessLogs bool          `yaml:"EnableAccessLogs" env:"ENABLE_ACCESS_LOGS"`
	EnableColorLogs  bool          `yaml:"EnableColorLogs" env:"ENABLE_COLOR_LOGS"`
	ShortedIDLength  int           `yaml:"ShortedIDLength" env:"SHORTED_ID_LENGTH"`
	Google           oAuthConf     `yaml:"Google" env:"GOOGLE"`
	GitHub           oAuthConf     `yaml:"GitHub" env:"GITHUB"`
	Microsoft        oAuthConf     `yaml:"Microsoft" env:"MICROSOFT"`
	Okta             oAuthConf     `yaml:"Okta" env:"OKTA"`
	GenericOIDC      oAuthConf     `yaml:"GenericOIDC" env:"GENERIC_OIDC"`
	Proxy            proxyAuthConf `yaml:"Proxy" env:"PROXY"`
	Redis            redisConf     `yaml:"Redis" env:"REDIS"`
}

type redisConf struct {
	Host         string `yaml:"Host" env:"HOST"`
	Password     string `yaml:"Password" env:"PASSWORD"`
	DB           int    `yaml:"DB" env:"DB"`
	MaxRetries   int    `yaml:"MaxRetries" env:"MAX_RETRIES"`
	ReadTimeout  string `yaml:"ReadTimeout" env:"READ_TIMEOUT"`
	WriteTimeout string `yaml:"WriteTimeout" env:"WRITE_TIMEOUT"`
	SessionDB    string `yaml:"SessionDB" env:"SESSION_DB"`
	SharedKey    string `yaml:"SharedKey" env:"SHARED_KEY"`
}

type oAuthConf struct {
	ClientID     string `yaml:"ClientID" env:"CLIENT_ID"`
	ClientSecret string `yaml:"ClientSecret" env:"CLIENT_SECRET"`
	EndpointURL  string `yaml:"EndpointURL" env:"ENDPOINT_URL"` // Optional for GitHub, mandatory for Okta and GenericOIDC
}

type proxyAuthConf struct {
	RequireUserHeader bool   `yaml:"RequireUserHeader" env:"REQUIRE_USER_HEADER"`
	UserHeader        string `yaml:"UserHeader" env:"USER_HEADER"`
	DisplayNameHeader string `yaml:"DisplayNameHeader" env:"DISPLAY_NAME_HEADER"`
}

// Config contains the default values
var Config = Configuration{
	ListenAddr:       ":8080",
	BaseURL:          "http://localhost:8080",
	DisplayURL:       "",
	DataDir:          "data",
	Backend:          "boltdb",
	EnableDebugMode:  false,
	EnableAccessLogs: true,
	EnableColorLogs:  true,
	UseSSL:           false,
	ShortedIDLength:  4,
	AuthBackend:      "oauth",
	Redis: redisConf{
		Host:         "127.0.0.1:6379",
		MaxRetries:   3,
		ReadTimeout:  "3s",
		WriteTimeout: "3s",
		SessionDB:    "1",
		SharedKey:    "secret",
	},
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
	logrus.Infof("Loaded configuration: %+v", Config)
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
	// if DisplayURL is not set in the config, default to BaseURL
	if Config.DisplayURL == "" {
		Config.DisplayURL = Config.BaseURL
	}

	return Config
}

// SetConfig sets the configuration into the memory
func SetConfig(c Configuration) {
	Config = c
}
