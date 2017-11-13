// Package store provides support to interact with the entries
package store

import (
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

type EntryPublicData struct {
	CreatedOn, LastVisit time.Time
	VisitCount           int
	URL                  string
}

// ErrNoEntryFound is returned when no entry to a id is found
var ErrNoEntryFound = errors.New("no entry found with this ID")

// ErrNoValidURL is returned when the URL is not valid
var ErrNoValidURL = errors.New("the given URL is no valid URL")

// ErrGeneratingIDFailed is returned when the 10 tries to generate an id failed
var ErrGeneratingIDFailed = errors.New("could not generate unique id, all ten tries failed")

// ErrIDIsEmpty is returned when the given ID is empty
var ErrIDIsEmpty = errors.New("the given ID is empty")

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
		idLength:   viper.GetInt("General.ShortedIDLength"),
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
	return entry, json.Unmarshal(raw, &entry)
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
	err = s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(s.bucketName).Put([]byte(id), raw); err != nil {
			return errors.Wrap(err, "could not put updated visitor count JSON into the bucket")
		}
		return nil
	})
	return err
}

// GetEntryByIDRaw returns the raw data (JSON) of a data set
func (s *Store) GetEntryByIDRaw(id string) ([]byte, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(s.bucketName).Get([]byte(id))
		if raw == nil {
			return ErrNoEntryFound
		}
		return nil
	})
	return raw, err
}

// CreateEntry creates a new record and returns his short id
func (s *Store) CreateEntry(entry Entry, givenID string) (string, error) {
	if !govalidator.IsURL(entry.Public.URL) {
		return "", ErrNoValidURL
	}
	// try it 10 times to make a short URL
	for i := 1; i <= 10; i++ {
		id, err := s.createEntry(entry, givenID)
		if err != nil && givenID != "" {
			return "", err
		} else if err != nil {
			logrus.Debugf("Could not create entry: %v", err)
			continue
		}
		return id, nil
	}
	return "", ErrGeneratingIDFailed
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.db.Close()
}
