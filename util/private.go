package util

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var privateKey []byte

// CheckForPrivateKey checks if already an private key exists, if not it will
// be randomly generated and saved as a private.dat file in the data directory
func CheckForPrivateKey() error {
	privateDatPath := filepath.Join(config.DataDir, "private.dat")
	privateDatContent, err := ioutil.ReadFile(privateDatPath)
	if err == nil {
		privateKey = privateDatContent
	} else if os.IsNotExist(err) {
		randomGeneratedKey := make([]byte, 256)
		if _, err = rand.Read(randomGeneratedKey); err != nil {
			return errors.Wrap(err, "could not read random bytes")
		}
		if err = ioutil.WriteFile(privateDatPath, randomGeneratedKey, 0644); err != nil {
			return errors.Wrap(err, "could not write private key")
		}
		privateKey = randomGeneratedKey
	} else if err != nil {
		return errors.Wrap(err, "could not read private key")
	}
	return nil
}

// GetPrivateKey returns the private key
func GetPrivateKey() []byte {
	return privateKey
}
