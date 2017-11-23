package util

import (
	"os"
	"testing"
)

func TestCheckforPrivateKey(t *testing.T) {
	TestReadInConfig(t)
	privateKey = nil
	if err := CheckForPrivateKey(); err != nil {
		t.Fatalf("could not check for private key: %v", err)
	}
	if GetPrivateKey() == nil {
		t.Fatalf("private key is nil")
	}
	if err := os.RemoveAll(GetDataDir()); err != nil {
		t.Fatalf("could not remove data dir: %v", err)
	}
}
