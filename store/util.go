package store

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"math/big"
	"time"
	"unicode"

	"github.com/boltdb/bolt"
	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/pkg/errors"
)

// createEntryRaw creates a entry with the given key value pair
func (s *Store) createEntryRaw(key, value, userIdentifier []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(shortedURLsBucket)
		if raw := bucket.Get(key); raw != nil {
			return errors.New("entry already exists")
		}
		if err := bucket.Put(key, value); err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		uTsURLsBucket, err := tx.CreateBucketIfNotExists(shortedIDsToUserBucket)
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return uTsURLsBucket.Put(key, userIdentifier)
	})
}

// createEntry creates a new entry with a randomly generated id. If on is present
// then the given ID is used
func (s *Store) createEntry(entry Entry, entryID string) (string, []byte, error) {
	var err error
	if entryID == "" {
		if entryID, err = generateRandomString(s.idLength); err != nil {
			return "", nil, errors.Wrap(err, "could not generate random string")
		}
	}
	entry.Public.CreatedOn = time.Now()
	rawEntry, err := json.Marshal(entry)
	if err != nil {
		return "", nil, err
	}
	mac := hmac.New(sha512.New, util.GetPrivateKey())
	if _, err := mac.Write([]byte(entryID)); err != nil {
		return "", nil, errors.Wrap(err, "could not write hmac")
	}
	return entryID, mac.Sum(nil), s.createEntryRaw([]byte(entryID), rawEntry, []byte(entry.OAuthProvider+entry.OAuthID))
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
