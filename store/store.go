// Package store provides support to interact with the entries
package store

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/pkg/errors"
)

// Store holds internal funcs and vars about the store
type Store struct {
	db         *bolt.DB
	bucketName []byte
	idLength   uint
	log        *logrus.Logger
}

// Entry is the data set which is stored in the DB as JSON
type Entry struct {
	URL                    string
	VisitCount             int
	RemoteAddr             string `json:",omitempty"`
	OAuthProvider, OAuthID string
	CreatedOn, LastVisit   time.Time
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
func New(storeConfig config.Store, log *logrus.Logger) (*Store, error) {
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
		log:        log,
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
	return entry, json.Unmarshal(raw, &entry)
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
		if err := bucket.Put([]byte(id), raw); err != nil {
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
func (s *Store) CreateEntry(URL, remoteAddr, oAuthProvider, oAuthID string) (string, error) {
	if !govalidator.IsURL(URL) {
		return "", ErrNoValidURL
	}
	// try it 10 times to make a short URL
	for i := 1; i <= 10; i++ {
		id, err := s.createEntry(URL, remoteAddr, oAuthProvider, oAuthID)
		if err != nil {
			s.log.Debugf("Could not create entry: %v", err)
			continue
		}
		return id, nil
	}
	return "", ErrGeneratingTriesFailed
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.db.Close()
}
