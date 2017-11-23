package main

import (
	"net/http"
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
	time.Sleep(1) // Give the http server a second to boot up
	if err := http.ListenAndServe(viper.GetString("listen_addr"), nil); err == nil {
		t.Fatal("port is not in use")
	}
	close()
	os.Exit(0)
}
