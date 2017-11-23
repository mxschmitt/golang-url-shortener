package util

import (
	"testing"

	"github.com/spf13/viper"
)

func TestReadInConfig(t *testing.T) {
	DoNotSetConfigName = true
	viper.SetConfigFile("test.yaml")
	if err := ReadInConfig(); err != nil {
		t.Fatalf("could not read in config file: %v", err)
	}
}
