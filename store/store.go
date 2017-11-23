// Package store provides support to interact with the entries
package store

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// Store holds internal funcs and vars about the store
type Store struct {
	db         *bolt.DB
	bucketName []byte
	idLength   int
}

// Entry is the data set which is stored in the DB as JSON
type Entry struct {
	OAuthProvider, OAuthID string
	RemoteAddr             string `json:",omitempty"`
	Public                 EntryPublicData
}

// EntryPublicData is the public part of an entry
type EntryPublicData struct {
	CreatedOn, LastVisit time.Time
	Expiration           *time.Time `json:",omitempty"`
	VisitCount           int
	URL                  string
}

// ErrNoEntryFound is returned when no entry to a id is found
var ErrNoEntryFound = errors.New("no entry found with this ID")

// ErrNoValidURL is returned when the URL is not valid
var ErrNoValidURL = errors.New("the given URL is no valid URL")

// ErrGeneratingIDFailed is returned when the 10 tries to generate an id failed
var ErrGeneratingIDFailed = errors.New("could not generate unique id, all ten tries failed")

// ErrEntryIsExpired is returned when the entry is expired
var ErrEntryIsExpired = errors.New("entry is expired")

// New initializes the store with the db
func New() (*Store, error) {
	db, err := bolt.Open(filepath.Join(util.GetDataDir(), "main.db"), 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open bolt DB database")
	}
	bucketName := []byte("shorted")
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Store{
		db:         db,
		idLength:   viper.GetInt("shorted_id_length"),
		bucketName: bucketName,
	}, nil
}

// GetEntryByID returns a unmarshalled entry of the db by a given ID
func (s *Store) GetEntryByID(id string) (*Entry, error) {
	if id == "" {
		return nil, ErrNoEntryFound
	}
	rawEntry, err := s.GetEntryByIDRaw(id)
	if err != nil {
		return nil, err
	}
	var entry *Entry
	return entry, json.Unmarshal(rawEntry, &entry)
}

// IncreaseVisitCounter increments the visit counter of an entry
func (s *Store) IncreaseVisitCounter(id string) error {
	entry, err := s.GetEntryByID(id)
	if err != nil {
		return errors.Wrap(err, "could not get entry by ID")
	}
	entry.Public.VisitCount++
	entry.Public.LastVisit = time.Now()
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(s.bucketName).Put([]byte(id), raw); err != nil {
			return errors.Wrap(err, "could not put updated visitor count JSON into the bucket")
		}
		return nil
	})
}

// GetURLAndIncrease Increases the visitor count, checks
// if the URL is expired and returns the origin URL
func (s *Store) GetURLAndIncrease(id string) (string, error) {
	entry, err := s.GetEntryByID(id)
	if err != nil {
		return "", err
	}
	if entry.Public.Expiration != nil && time.Now().After(*entry.Public.Expiration) {
		return "", ErrEntryIsExpired
	}
	if err := s.IncreaseVisitCounter(id); err != nil {
		return "", errors.Wrap(err, "could not increase visitor counter")
	}
	return entry.Public.URL, nil
}

// GetEntryByIDRaw returns the raw data (JSON) of a data set
func (s *Store) GetEntryByIDRaw(id string) ([]byte, error) {
	var raw []byte
	return raw, s.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(s.bucketName).Get([]byte(id))
		if raw == nil {
			return ErrNoEntryFound
		}
		return nil
	})
}

// CreateEntry creates a new record and returns his short id
func (s *Store) CreateEntry(entry Entry, givenID string) (string, []byte, error) {
	if !govalidator.IsURL(entry.Public.URL) {
		return "", nil, ErrNoValidURL
	}
	// try it 10 times to make a short URL
	for i := 1; i <= 10; i++ {
		id, delID, err := s.createEntry(entry, givenID)
		if err != nil && givenID != "" {
			return "", nil, err
		} else if err != nil {
			logrus.Debugf("Could not create entry: %v", err)
			continue
		}
		return id, delID, nil
	}
	return "", nil, ErrGeneratingIDFailed
}

// DeleteEntry deletes an Entry fully from the DB
func (s *Store) DeleteEntry(id string, givenHmac []byte) error {
	mac := hmac.New(sha512.New, util.GetPrivateKey())
	if _, err := mac.Write([]byte(id)); err != nil {
		return errors.Wrap(err, "could not write hmac")
	}
	if !hmac.Equal(mac.Sum(nil), givenHmac) {
		return errors.New("hmac verification failed")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		if bucket.Get([]byte(id)) == nil {
			return errors.New("entry already deleted")
		}
		return bucket.Delete([]byte(id))
	})
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.db.Close()
}
