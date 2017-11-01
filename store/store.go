// Package store provides support to interact with the entries
package store

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/maxibanki/golang-url-shortener/config"
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
	URL        string
	VisitCount int
	RemoteAddr string `json:",omitempty"`
	CreatedOn  time.Time
	LastVisit  time.Time
}

// ErrNoEntryFound is returned when no entry to a id is found
var ErrNoEntryFound = errors.New("no entry found")

// ErrNoValidURL is returned when the URL is not valid
var ErrNoValidURL = errors.New("no valid URL")

// ErrGeneratingTriesFailed is returned when the 10 tries to generate an id failed
var ErrGeneratingTriesFailed = errors.New("could not generate unique id, db full?")

// ErrIDIsEmpty is returned when the given ID is empty
var ErrIDIsEmpty = errors.New("id is empty")

// New initializes the store with the db
func New(storeConfig config.Store) (*Store, error) {
	db, err := bolt.Open(storeConfig.DBPath, 0644, &bolt.Options{Timeout: 1 * time.Second})
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
		idLength:   storeConfig.ShortedIDLength,
		bucketName: bucketName,
	}, nil
}

// GetEntryByID returns a unmarshalled entry of the db by a given ID
func (s *Store) GetEntryByID(id string) (*Entry, error) {
	if id == "" {
		return nil, ErrIDIsEmpty
	}
	raw, err := s.GetEntryByIDRaw(id)
	if err != nil {
		return nil, err
	}
	var entry *Entry
	err = json.Unmarshal(raw, &entry)
	return entry, err
}

// IncreaseVisitCounter increments the visit counter of an entry
func (s *Store) IncreaseVisitCounter(id string) error {
	entry, err := s.GetEntryByID(id)
	if err != nil {
		return err
	}
	entry.VisitCount++
	entry.LastVisit = time.Now()
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		err := bucket.Put([]byte(id), raw)
		if err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		return nil
	})
	return err
}

// GetEntryByIDRaw returns the raw data (JSON) of a data set
func (s *Store) GetEntryByIDRaw(id string) ([]byte, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		raw = bucket.Get([]byte(id))
		if raw == nil {
			return ErrNoEntryFound
		}
		return nil
	})
	return raw, err
}

// CreateEntry creates a new record and returns his short id
func (s *Store) CreateEntry(URL, remoteAddr string) (string, error) {
	if !govalidator.IsURL(URL) {
		return "", ErrNoValidURL
	}
	// try it 10 times to make a short URL
	for i := 1; i <= 10; i++ {
		id, err := s.createEntry(URL, remoteAddr)
		if err != nil {
			continue
		}
		return id, nil
	}
	return "", ErrGeneratingTriesFailed
}

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

// createEntryRaw creates a entry with the given key value pair
func (s *Store) createEntryRaw(key, value []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		raw := bucket.Get(key)
		if raw != nil {
			return errors.New("entry value is not empty")
		}
		err := bucket.Put(key, value)
		if err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		return nil
	})
	return err
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.db.Close()
}

// generateRandomString generates a random string with an predefined length
func generateRandomString(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
