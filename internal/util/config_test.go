package util

import (
	"testing"
)

func TestReadInConfig(t *testing.T) {
	if err := ReadInConfig(); err != nil {
		t.Fatalf("could not read in config file: %v", err)
	}
	config := Config
	config.DataDir = "./test"
}
