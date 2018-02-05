package main

import (
	"net"
	"testing"
	"time"

	"github.com/mxschmitt/golang-url-shortener/util"
)

func TestInitShortener(t *testing.T) {
	close, err := initShortener()
	if err != nil {
		t.Fatalf("could not init shortener: %v", err)
	}
	time.Sleep(time.Millisecond * 200) // Give the http server a second to boot up
	// We expect there a port is in use error
	if _, err := net.Listen("tcp", util.GetConfig().ListenAddr); err == nil {
		t.Fatalf("port is not in use: %v", err)
	}
	close()
}
