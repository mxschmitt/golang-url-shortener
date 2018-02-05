// Package stores provides support to interact with the entries
package stores

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"math/big"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/asaskevich/govalidator"
	"github.com/mxschmitt/golang-url-shortener/stores/boltdb"
	"github.com/mxschmitt/golang-url-shortener/stores/shared"
	"github.com/mxschmitt/golang-url-shortener/util"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Store holds internal funcs and vars about the store
type Store struct {
	storage  shared.Storage
	idLength int
}

// ErrNoValidURL is returned when the URL is not valid
var ErrNoValidURL = errors.New("the given URL is no valid URL")

// ErrGeneratingIDFailed is returned when the 10 tries to generate an id failed
var ErrGeneratingIDFailed = errors.New("could not generate unique id, all ten tries failed")

// ErrEntryIsExpired is returned when the entry is expired
var ErrEntryIsExpired = errors.New("entry is expired")

// New initializes the store with the db
func New() (*Store, error) {
	s, err := boltdb.New(filepath.Join(util.GetConfig().DataDir, "main.db"))
	if err != nil {
		return nil, errors.Wrap(err, "could not create bolt db store")
	}
	return &Store{
		storage:  s,
		idLength: util.GetConfig().ShortedIDLength,
	}, nil
}

// GetEntryByID returns a unmarshalled entry of the db by a given ID
func (s *Store) GetEntryByID(id string) (*shared.Entry, error) {
	if id == "" {
		return nil, shared.ErrNoEntryFound
	}
	return s.storage.GetEntryByID(id)
}

// GetEntryAndIncrease Increases the visitor count, checks
// if the URL is expired and returns the origin URL
func (s *Store) GetEntryAndIncrease(id string) (*shared.Entry, error) {
	entry, err := s.GetEntryByID(id)
	if err != nil {
		return nil, err
	}
	if entry.Public.Expiration != nil && time.Now().After(*entry.Public.Expiration) {
		return nil, ErrEntryIsExpired
	}
	if err := s.storage.IncreaseVisitCounter(id); err != nil {
		return nil, errors.Wrap(err, "could not increase visitor counter")
	}
	entry.Public.VisitCount++
	return entry, nil
}

// CreateEntry creates a new record and returns his short id
func (s *Store) CreateEntry(entry shared.Entry, givenID, password string) (string, []byte, error) {
	entry.Public.URL = strings.Replace(entry.Public.URL, " ", "%20", -1)
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
		id, passwordHash, err := s.createEntry(entry, givenID)
		if err != nil && givenID != "" {
			return "", nil, err
		} else if err != nil {
			logrus.Debugf("Could not create entry: %v", err)
			continue
		}
		return id, passwordHash, nil
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
	return errors.Wrap(s.storage.DeleteEntry(id), "could not delete entry")
}

// RegisterVisit registers an new incoming request in the store
func (s *Store) RegisterVisit(id string, visitor shared.Visitor) {
	requestID := uuid.New()
	logrus.WithFields(logrus.Fields{
		"ClientIP":  visitor.IP,
		"ID":        id,
		"RequestID": requestID,
	}).Info("New redirect was registered...")
	if err := s.storage.RegisterVisitor(id, requestID, visitor); err != nil {
		logrus.Warningf("could not register visit: %v", err)
	}
}

// GetVisitors returns all the visits of a shorted URL
func (s *Store) GetVisitors(id string) ([]shared.Visitor, error) {
	visitors, err := s.storage.GetVisitors(id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get visitors")
	}
	return visitors, nil
}

// GetUserEntries returns all the shorted URL entries of an user
func (s *Store) GetUserEntries(oAuthProvider, oAuthID string) (map[string]shared.Entry, error) {
	userIdentifier := getUserIdentifier(oAuthProvider, oAuthID)
	entries, err := s.storage.GetUserEntries(userIdentifier)
	if err != nil {
		return nil, errors.Wrap(err, "could not get user entries")
	}
	return entries, nil
}

func getUserIdentifier(oAuthProvider, oAuthID string) string {
	return oAuthProvider + oAuthID
}

// Close closes the bolt db database
func (s *Store) Close() error {
	return s.storage.Close()
}

// createEntry creates a new entry with a randomly generated id. If on is present
// then the given ID is used
func (s *Store) createEntry(entry shared.Entry, entryID string) (string, []byte, error) {
	var err error
	if entryID == "" {
		if entryID, err = generateRandomString(s.idLength); err != nil {
			return "", nil, errors.Wrap(err, "could not generate random string")
		}
	}
	entry.Public.CreatedOn = time.Now()
	mac := hmac.New(sha512.New, util.GetPrivateKey())
	if _, err := mac.Write([]byte(entryID)); err != nil {
		return "", nil, errors.Wrap(err, "could not write hmac")
	}
	if err := s.storage.CreateEntry(entry, entryID, getUserIdentifier(entry.OAuthProvider, entry.OAuthID)); err != nil {
		return "", nil, errors.Wrap(err, "could not create entry")
	}
	return entryID, mac.Sum(nil), nil
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
