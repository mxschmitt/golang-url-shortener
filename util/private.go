package util

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var privateKey []byte

// CheckForPrivateKey checks if already an private key exists, if not one will be randomly generated
func CheckForPrivateKey() error {
	privateDat := filepath.Join(GetDataDir(), "private.dat")
	d, err := ioutil.ReadFile(privateDat)
	if err == nil {
		privateKey = d
	} else if os.IsNotExist(err) {
		b := make([]byte, 256)
		if _, err := rand.Read(b); err != nil {
			return errors.Wrap(err, "could not read random bytes")
		}
		if err = ioutil.WriteFile(privateDat, b, 0644); err != nil {
			return errors.Wrap(err, "could not write private key")
		}
		privateKey = b
	} else if err != nil {
		return errors.Wrap(err, "could not read private key")
	}
	return nil
}

// GetPrivateKey returns the private key from the memory
func GetPrivateKey() []byte {
	return privateKey
}
