package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestInitShortener(t *testing.T) {
	close, err := initShortener()
	if err != nil {
		t.Fatalf("could not init shortener: %v", err)
	}
	time.Sleep(time.Second) // Give the http server a second to boot up
	// We expect there a port is in use error
	if _, err := net.Listen("tcp", viper.GetString("listen_addr")); err == nil {
		t.Fatalf("port is not in use: %v", err)
	}
	close()
	os.Exit(0)
}
