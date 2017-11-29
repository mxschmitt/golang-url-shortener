// Package store provides support to interact with the entries
package store

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// Store holds internal funcs and vars about the store
type Store struct {
	db       *bolt.DB
	idLength int
}

// Entry is the data set which is stored in the DB as JSON
type Entry struct {
	OAuthProvider, OAuthID string
	RemoteAddr             string `json:",omitempty"`
	DeletionURL            string `json:",omitempty"`
	Password               []byte `json:",omitempty"`
	Public                 EntryPublicData
}

// Visitor is the entry which is stored in the visitors bucket
type Visitor struct {
	IP, Referer, UserAgent                                 string
	Timestamp                                              time.Time
	UTMSource, UTMMedium, UTMCampaign, UTMContent, UTMTerm string `json:",omitempty"`
}

// EntryPublicData is the public part of an entry
type EntryPublicData struct {
	CreatedOn             time.Time
	LastVisit, Expiration *time.Time `json:",omitempty"`
	VisitCount            int
	URL                   string
}

// ErrNoEntryFound is returned when no entry to a id is found
var ErrNoEntryFound = errors.New("no entry found with this ID")

// ErrNoValidURL is returned when the URL is not valid
var ErrNoValidURL = errors.New("the given URL is no valid URL")

// ErrGeneratingIDFailed is returned when the 10 tries to generate an id failed
var ErrGeneratingIDFailed = errors.New("could not generate unique id, all ten tries failed")

// ErrEntryIsExpired is returned when the entry is expired
var ErrEntryIsExpired = errors.New("entry is expired")

var (
	shortedURLsBucket      = []byte("shorted")
	shortedIDsToUserBucket = []byte("shorted2Users")
)

// New initializes the store with the db
func New() (*Store, error) {
	db, err := bolt.Open(filepath.Join(util.GetConfig().DataDir, "main.db"), 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open bolt DB database")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(shortedURLsBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Store{
		db:       db,
		idLength: util.GetConfig().ShortedIDLength,
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
	currentTime := time.Now()
	entry.Public.LastVisit = &currentTime
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(shortedURLsBucket).Put([]byte(id), raw); err != nil {
			return errors.Wrap(err, "could not put updated visitor count JSON into the bucket")
		}
		return nil
	})
}

// GetEntryAndIncrease Increases the visitor count, checks
// if the URL is expired and returns the origin URL
func (s *Store) GetEntryAndIncrease(id string) (*Entry, error) {
	entry, err := s.GetEntryByID(id)
	if err != nil {
		return nil, err
	}
	if entry.Public.Expiration != nil && time.Now().After(*entry.Public.Expiration) {
		return nil, ErrEntryIsExpired
	}
	if err := s.IncreaseVisitCounter(id); err != nil {
		return nil, errors.Wrap(err, "could not increase visitor counter")
	}
	return entry, nil
}

// GetEntryByIDRaw returns the raw data (JSON) of a data set
func (s *Store) GetEntryByIDRaw(id string) ([]byte, error) {
	var raw []byte
	return raw, s.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(shortedURLsBucket).Get([]byte(id))
		if raw == nil {
			return ErrNoEntryFound
		}
		return nil
	})
}

// CreateEntry creates a new record and returns his short id
func (s *Store) CreateEntry(entry Entry, givenID, password string) (string, []byte, error) {
	if !govalidator.IsURL(entry.Public.URL) {
		return "", nil, ErrNoValidURL
	}
	if password != "" {
		var err error
		entry.Password, err = bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return "", nil, errors.Wrap(err, "could not generate bcrypt from password")
		}
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
		bucket := tx.Bucket(shortedURLsBucket)
		if bucket.Get([]byte(id)) == nil {
			return errors.New("entry already deleted")
		}
		if err := bucket.Delete([]byte(id)); err != nil {
			return errors.Wrap(err, "could not delete entry")
		}
		if err := tx.DeleteBucket([]byte(id)); err != nil && err != bolt.ErrBucketNotFound && err != bolt.ErrBucketExists {
			return errors.Wrap(err, "could not delte bucket")
		}
		uTsIDsBucket := tx.Bucket(shortedIDsToUserBucket)
		return uTsIDsBucket.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, []byte(id)) {
				return uTsIDsBucket.Delete(k)
			}
			return nil
		})
	})
}

// RegisterVisit registers an new incoming request in the store
func (s *Store) RegisterVisit(id string, visitor Visitor) {
	requestID := uuid.New()
	logrus.WithFields(logrus.Fields{
		"ClientIP":  visitor.IP,
		"ID":        id,
		"RequestID": requestID,
	}).Info("New redirect was registered...")

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		data, err := json.Marshal(visitor)
		if err != nil {
			return errors.Wrap(err, "could not create json")
		}
		return bucket.Put([]byte(requestID), data)
	})
	if err != nil {
		logrus.Warningf("could not register visit: %v", err)
	}
}

// GetVisitors returns all the visits of a shorted URL
func (s *Store) GetVisitors(id string) ([]Visitor, error) {
	output := []Visitor{}
	return output, s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return bucket.ForEach(func(k, v []byte) error {
			var value Visitor
			if err := json.Unmarshal(v, &value); err != nil {
				return errors.Wrap(err, "could not unmarshal json")
			}
			output = append(output, value)
			return nil
		})
	})
}

// GetUserEntries returns all the shorted URL entries of an user
func (s *Store) GetUserEntries(oAuthProvider, oAuthID string) (map[string]Entry, error) {
	entries := map[string]Entry{}
	return entries, s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(shortedIDsToUserBucket)
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return bucket.ForEach(func(k, v []byte) error {
			if bytes.Equal(v, []byte(oAuthProvider+oAuthID)) {
				entry, err := s.GetEntryByID(string(k))
				if err != nil {
					return errors.Wrap(err, "could not get entry")
				}
				entries[string(k)] = *entry
			}
			return nil
		})
	})
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.db.Close()
}
