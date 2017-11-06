package store

import (
	"encoding/json"
	"math/rand"
	"time"

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
func (s *Store) createEntry(URL, remoteAddr string) (string, error) {
	id := generateRandomString(s.idLength)
	exists := s.checkExistence(id)
	if !exists {
		raw, err := json.Marshal(Entry{
			URL:        URL,
			RemoteAddr: remoteAddr,
			CreatedOn:  time.Now(),
		})
		if err != nil {
			return "", err
		}
		return id, s.createEntryRaw([]byte(id), raw)
	}
	return "", errors.New("entry already exists")
}

// generateRandomString generates a random string with an predefined length
func generateRandomString(length uint) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
