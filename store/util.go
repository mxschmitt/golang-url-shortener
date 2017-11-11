package store

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"time"
	"unicode"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// checkExistens returns true if a entry with a given ID
// exists and false if not
func (s *Store) checkExistence(id string) bool {
	raw, err := s.GetEntryByIDRaw(id)
	if err != nil && err != ErrNoEntryFound {
		return true
	}
	if raw != nil {
		return true
	}
	return false
}

// createEntryRaw creates a entry with the given key value pair
func (s *Store) createEntryRaw(key, value []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		raw := bucket.Get(key)
		if raw != nil {
			return errors.New("entry value is not empty")
		}
		if err := bucket.Put(key, value); err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		return nil
	})
	return err
}

// createEntry creates a new entry
func (s *Store) createEntry(entry Entry) (string, error) {
	id, err := generateRandomString(s.idLength)
	if err != nil {
		return "", errors.Wrap(err, "could not generate random string")
	}
	exists := s.checkExistence(id)
	if !exists {
		entry.CreatedOn = time.Now()
		raw, err := json.Marshal(entry)
		if err != nil {
			return "", err
		}
		return id, s.createEntryRaw([]byte(id), raw)
	}
	return "", errors.New("entry already exists")
}

// generateRandomString generates a random string with an predefined length
func generateRandomString(length uint) (string, error) {
	var result string
	for len(result) < int(length) {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}
		n := num.Int64()
		if unicode.IsLetter(rune(n)) {
			result += string(n)
		}
	}
	return result, nil
}
