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

// createEntryRaw creates a entry with the given key value pair
func (s *Store) createEntryRaw(key, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		if raw := bucket.Get(key); raw != nil {
			return errors.New("entry already exists")
		}
		if err := bucket.Put(key, value); err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		return nil
	})
}

// createEntry creates a new entry with a randomly generated id. If on is present
// then the given ID is used
func (s *Store) createEntry(entry Entry, entryID string) (string, error) {
	var err error
	if entryID == "" {
		if entryID, err = generateRandomString(s.idLength); err != nil {
			return "", errors.Wrap(err, "could not generate random string")
		}
	}
	entry.Public.CreatedOn = time.Now()
	rawEntry, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return entryID, s.createEntryRaw([]byte(entryID), rawEntry)
}

// generateRandomString generates a random string with an predefined length
func generateRandomString(length int) (string, error) {
	var result string
	for len(result) < length {
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
