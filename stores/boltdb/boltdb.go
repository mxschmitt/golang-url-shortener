package boltdb

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mxschmitt/golang-url-shortener/stores/shared"
	"github.com/pkg/errors"
)

var (
	shortedURLsBucket      = []byte("shorted")
	shortedIDsToUserBucket = []byte("shorted2Users")
)

// BoltStore implements the stores.Storage interface
type BoltStore struct {
	db *bolt.DB
}

// New returns a bolt store which implements the stores.Storage interface
func New(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "could not open bolt DB database")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(shortedURLsBucket); err != nil {
			return errors.Wrapf(err, "could not create %s bucket", shortedURLsBucket)
		}
		if _, err := tx.CreateBucketIfNotExists(shortedIDsToUserBucket); err != nil {
			return errors.Wrapf(err, "could not create %s bucket", shortedIDsToUserBucket)
		}
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create buckets")
	}
	return &BoltStore{
		db: db,
	}, nil
}

// Close closes the bolt database
func (b *BoltStore) Close() error {
	return b.db.Close()
}

// GetEntryByID returns a entry and an error by the shorted ID
func (b *BoltStore) GetEntryByID(id string) (*shared.Entry, error) {
	var raw []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		raw = tx.Bucket(shortedURLsBucket).Get([]byte(id))
		if raw == nil {
			return shared.ErrNoEntryFound
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not view db")
	}
	var entry *shared.Entry
	return entry, json.Unmarshal(raw, &entry)
}

// IncreaseVisitCounter increases the visit counter and sets the current
// time as the last visit ones
func (b *BoltStore) IncreaseVisitCounter(id string) error {
	entry, err := b.GetEntryByID(id)
	if err != nil {
		return errors.Wrap(err, "could not get entry by ID")
	}
	entry.Public.VisitCount++
	currentTime := time.Now()
	entry.Public.LastVisit = &currentTime
	raw, err := json.Marshal(entry)
	if err != nil {
		return errors.Wrap(err, "could not marshal json")
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(shortedURLsBucket).Put([]byte(id), raw); err != nil {
			return errors.Wrap(err, "could not put updated visitor")
		}
		return nil
	})
	return errors.Wrap(err, "could not update entry")
}

// CreateEntry creates an entry by a given ID and returns an error
func (b *BoltStore) CreateEntry(entry shared.Entry, id, userIdentifier string) error {
	entryRaw, err := json.Marshal(entry)
	if err != nil {
		return errors.Wrap(err, "could not marshal entry")
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(shortedURLsBucket)
		if raw := bucket.Get([]byte(id)); raw != nil {
			return errors.New("entry already exists")
		}
		if err := bucket.Put([]byte(id), entryRaw); err != nil {
			return errors.Wrap(err, "could not put data into bucket")
		}
		uTsURLsBucket, err := tx.CreateBucketIfNotExists(shortedIDsToUserBucket)
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return uTsURLsBucket.Put([]byte(id), []byte(userIdentifier))
	})
	return errors.Wrap(err, "could not update db")
}

// DeleteEntry deleted an entry by a given ID and returns an error
func (b *BoltStore) DeleteEntry(id string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
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
	return errors.Wrap(err, "could not update db")
}

// GetVisitors returns the visitors and an error of an entry
func (b *BoltStore) GetVisitors(id string) ([]shared.Visitor, error) {
	output := []shared.Visitor{}
	return output, b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return bucket.ForEach(func(k, v []byte) error {
			var value shared.Visitor
			if err := json.Unmarshal(v, &value); err != nil {
				return errors.Wrap(err, "could not unmarshal json")
			}
			output = append(output, value)
			return nil
		})
	})
}

// GetUserEntries returns all user entries of an given user identifier
func (b *BoltStore) GetUserEntries(userIdentifier string) (map[string]shared.Entry, error) {
	entries := map[string]shared.Entry{}
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(shortedIDsToUserBucket)
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		return bucket.ForEach(func(k, v []byte) error {
			if bytes.Equal(v, []byte(userIdentifier)) {
				entry, err := b.GetEntryByID(string(k))
				if err != nil {
					return errors.Wrap(err, "could not get entry")
				}
				entries[string(k)] = *entry
			}
			return nil
		})
	})
	return entries, errors.Wrap(err, "could not update db")
}

// RegisterVisitor saves the visitor in the database
func (b *BoltStore) RegisterVisitor(id, visitID string, visitor shared.Visitor) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return errors.Wrap(err, "could not create bucket")
		}
		data, err := json.Marshal(visitor)
		if err != nil {
			return errors.Wrap(err, "could not create json")
		}
		return bucket.Put([]byte(visitID), data)
	})
	return errors.Wrap(err, "could not update db")
}
