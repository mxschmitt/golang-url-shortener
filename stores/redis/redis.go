package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mxschmitt/golang-url-shortener/stores/shared"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	entryPathPrefix     = "entry:"       // prefix for path-to-url mappings
	entryUserPrefix     = "user:"        // prefix for path-to-user mappings
	userToEntriesPrefix = "userEntries:" // prefix for user-to-[]entries mappings (redis SET)
	entryVisitsPrefix   = "entryVisits:" // prefix for entry-to-[]visit mappings (redis LIST)
)

// Store implements the stores.Storage interface
type Store struct {
	c *redis.Client
}

// New initializes connection to the redis instance.
func New(hostaddr, password string, db int, maxRetries int, readTimeout string, writeTimeout string) (*Store, error) {
	var rt, wt time.Duration
	var err error

	if rt, err = time.ParseDuration(readTimeout); err != nil {
		return nil, errors.Wrap(err, "Could not parse read timeout")
	}
	if wt, err = time.ParseDuration(writeTimeout); err != nil {
		return nil, errors.Wrap(err, "Could not parse write timeout")
	}
	c := redis.NewClient(&redis.Options{
		Addr:         hostaddr,
		Password:     password,
		DB:           db,
		MaxRetries:   maxRetries,
		ReadTimeout:  rt,
		WriteTimeout: wt,
	})
	// if we can't talk to redis, fail fast
	if _, err = c.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "Could not connect to redis db0")
	}
	ret := &Store{c: c}
	return ret, nil
}

// keyExists checks for the existence of a key in redis.
func (r *Store) keyExists(key string) (exists bool, err error) {
	logrus.Debugf("Checking for existence of key: %s", key)
	result := r.c.Exists(key)
	if result.Err() != nil {
		msg := fmt.Sprintf("Error looking up key '%s': '%v', got val: '%d'", key, result.Err(), result.Val())
		logrus.Error(msg)
		return false, errors.Wrap(result.Err(), msg)
	}
	if result.Val() == 1 {
		logrus.Debugf("Key '%s' exists!", key)
		return true, nil
	}
	logrus.Debugf("Key '%s' does not exist!", key)
	return false, nil
}

// setValue sets the value of a key in redis.
func (r *Store) setValue(key string, raw []byte) error {
	logrus.Debugf("Setting value for key '%s: '%s''", key, raw)
	status := r.c.Set(key, raw, 0) // n.b. expiration 0 means never expire
	if status.Err() != nil {
		msg := fmt.Sprintf("Got an unexpected error adding key '%s': %s", key, status.Err())
		logrus.Error(msg)
		return errors.Wrap(status.Err(), msg)
	}
	return nil
}

// createValue is a wrapper around setValue that returns an error if the key already exists.
func (r *Store) createValue(key string, raw []byte) error {
	logrus.Debugf("Creating key '%s'", key)
	exists, err := r.keyExists(key)
	if err != nil {
		msg := fmt.Sprintf("Could not check existence of key '%s': %s", key, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	if exists == true {
		msg := fmt.Sprintf("Could not create key '%s':  already exists", key)
		logrus.Error(msg)
		return errors.New(msg)
	}
	return r.setValue(key, raw)
}

// delValue deletes a key in redis.
func (r *Store) delValue(key string) error {
	logrus.Debugf("Deleting key '%s'", key)

	exists, err := r.keyExists(key)
	if err != nil {
		msg := fmt.Sprintf("Could not check existence of key '%s': %s", key, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	if exists == false {
		logrus.Warnf("Tried to delete key '%s' but it's already gone", key)
		return err
	}

	status := r.c.Del(key)
	if status.Err() != nil {
		msg := fmt.Sprintf("Got an unexpected error deleting key '%s': %s", key, status.Err())
		logrus.Error(msg)
		return errors.Wrap(status.Err(), msg)
	}
	return err
}

// CreateEntry creates an entry (path->url mapping) and all associated stored data.
func (r *Store) CreateEntry(entry shared.Entry, id, userIdentifier string) error {
	// add the entry (path->url mapping)
	logrus.Debugf("Creating entry '%s' for user '%s'", id, userIdentifier)
	raw, err := json.Marshal(entry)
	if err != nil {
		msg := fmt.Sprintf("Could not marshal JSON for entry %s: %v", id, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	entryKey := entryPathPrefix + id
	logrus.Debugf("Adding key '%s': %s", entryKey, raw)
	err = r.createValue(entryKey, raw)
	if err != nil {
		msg := fmt.Sprintf("Failed to set key '%s' for user '%s': %v", entryKey, userIdentifier, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	// add the path->user mapping
	userKey := entryUserPrefix + id
	logrus.Debugf("Adding key '%s': %s", userKey, raw)
	err = r.createValue(userKey, []byte(userIdentifier))
	if err != nil {
		msg := fmt.Sprintf("Failed to set key '%s' for user '%s': %v", userKey, userIdentifier, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	// add the entry to the SET of entries for the useridentifier
	userEntriesKey := userToEntriesPrefix + userIdentifier
	logrus.Debugf("Adding entry '%s' to set of entries for user '%s'", id, userIdentifier)
	result := r.c.SAdd(userEntriesKey, id)
	if result.Err() != nil {
		msg := fmt.Sprintf("Failed to add entry '%s' for user '%s': %v", id, userIdentifier, result.Err())
		logrus.Error(msg)
		return errors.Wrap(result.Err(), msg)
	}
	logrus.Debugf("Successfully added entry '%s' to set '%s'", id, userEntriesKey)
	return nil
}

// DeleteEntry deletes an entry and all associated stored data.
func (r *Store) DeleteEntry(id string) error {
	// delete the id-to-url mapping
	entryKey := entryPathPrefix + id
	err := r.delValue(entryKey)
	if err != nil {
		msg := fmt.Sprintf("Could not delete entry id %s: %v", id, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	// delete the visitors list for the id
	entryVisitsKey := entryVisitsPrefix + id
	err = r.delValue(entryVisitsKey)
	if err != nil {
		msg := fmt.Sprintf("Could not delete visitors list for id %s: %v", id, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	// get the user for the id
	userKey := entryUserPrefix + id
	var userIdentifier string
	userIdentifier, err = r.c.Get(userKey).Result()
	if err != nil {
		msg := fmt.Sprintf("Could not fetch id to user mapping for id '%s': %v", id, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	// delete the entry from set of entries for the user
	userEntriesKey := userToEntriesPrefix + userIdentifier
	err = r.c.SRem(userEntriesKey, id).Err()
	if err != nil {
		msg := fmt.Sprintf("Could not remove entry '%s' from list of entries for user '%s': %v", id, userIdentifier, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	// delete the id-to-user mapping
	err = r.delValue(userKey)
	if err != nil {
		msg := fmt.Sprintf("Could not delete the path-to-user mapping for entry '%s': %v", id, err)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}

	return err
}

// GetEntryByID looks up an entry by its path and returns a pointer to a
// shared.Entry instance, with the visit count and last visit time set
// properly.
func (r *Store) GetEntryByID(id string) (*shared.Entry, error) {
	entryKey := entryPathPrefix + id
	logrus.Debugf("Fetching key: '%s'", entryKey)
	result := r.c.Get(entryKey)
	raw, err := result.Bytes()
	if err != nil {
		msg := fmt.Sprintf("Error looking up key '%s': %s'", entryKey, err)
		logrus.Warn(msg)
		err = shared.ErrNoEntryFound
		return nil, err
	}
	logrus.Debugf("Got entry for key '%s': '%s'", entryKey, raw)

	var entry *shared.Entry
	err = json.Unmarshal(raw, &entry)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling JSON for entry '%s': %v  (json str: '%s')", id, err, raw)
		logrus.Error(msg)
		return nil, errors.Wrap(err, msg)
	}

	// now we interleave the visit count and the last visit time
	// from the redis sources (we do this so we don't have to rewrite
	// the entry every time someone visits which is madness)
	//
	// first, the visit count is just the length of the visitors list
	entryVisitsKey := entryVisitsPrefix + id
	visitCount, err := r.c.LLen(entryVisitsKey).Result()
	if err != nil {
		logrus.Warnf("Could not get length of visitor list for id '%s': '%v'", id, err)
		entry.Public.VisitCount = int(0) // or zero if nobody's visited, that's fine.
	} else {
		entry.Public.VisitCount = int(visitCount)
	}

	// grab the timestamp out of the last visitor on the list
	var visitor *shared.Visitor
	lastVisit := time.Time(time.Unix(0, 0)) // default to start-of-epoch if we can't figure it out
	raw, err = r.c.LIndex(entryVisitsKey, 0).Bytes()
	if err != nil {
		logrus.Warnf("Could not fetch visitor list for entry '%s': %v", id, err)
	} else {
		err = json.Unmarshal(raw, &visitor)
		if err != nil {
			logrus.Warnf("Could not unmarshal JSON for last visitor to entry '%s': %v  (got string: '%s')", id, err, raw)
		} else {
			lastVisit = visitor.Timestamp
		}
	}
	logrus.Debugf("Setting last visit time for entry '%s' to '%v'", id, lastVisit)
	entry.Public.LastVisit = &lastVisit

	return entry, nil
}

// GetUserEntries returns all entries that are owned by a given user, in the
// form of a map of path->shared.Entry
func (r *Store) GetUserEntries(userIdentifier string) (map[string]shared.Entry, error) {
	logrus.Debugf("Getting all entries for user %s", userIdentifier)
	entries := map[string]shared.Entry{}
	key := userToEntriesPrefix + userIdentifier
	result := r.c.SMembers(key)
	if result.Err() != nil {
		msg := fmt.Sprintf("Could not fetch set of entries for user '%s': %v", userIdentifier, result.Err())
		logrus.Errorf(msg)
		return nil, errors.Wrap(result.Err(), msg)
	}
	for _, v := range result.Val() {
		logrus.Debugf("got entry: %s", v)
		entry, err := r.GetEntryByID(string(v))
		if err != nil {
			msg := fmt.Sprintf("Could not get entry '%s': %s", v, err)
			logrus.Warn(msg)
		} else {
			entries[string(v)] = *entry
		}
	}
	logrus.Debugf("all out of entries")
	return entries, nil
}

// RegisterVisitor adds a shared.Visitor to the list of visits for a path.
func (r *Store) RegisterVisitor(id, visitID string, visitor shared.Visitor) error {
	data, err := json.Marshal(visitor)
	if err != nil {
		msg := fmt.Sprintf("Could not marshal JSON for entry %s, visitID %s", id, visitID)
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	// push the visit data onto a redis list who's key is the url id
	key := entryVisitsPrefix + id
	result := r.c.LPush(key, data)
	if result.Err() != nil {
		msg := fmt.Sprintf("Could not register visitor for ID %s", id)
		logrus.Error(msg)
		return errors.Wrap(result.Err(), msg)
	}
	return err
}

// GetVisitors returns the full list of visitors for a path.
func (r *Store) GetVisitors(id string) ([]shared.Visitor, error) {
	var visitors []shared.Visitor
	key := entryVisitsPrefix + id
	// TODO: for non-trivial numbers of keys, this could start
	// to get hairy; should convert to a paginated Scan operation.
	result := r.c.LRange(key, 0, -1)
	if result.Err() != nil {
		msg := fmt.Sprintf("Could not get visitors for id '%s'", id)
		logrus.Error(msg)
		return nil, errors.Wrap(result.Err(), msg)
	}
	for _, v := range result.Val() {
		var value shared.Visitor
		if err := json.Unmarshal([]byte(v), &value); err != nil {
			msg := fmt.Sprintf("Could not unmarshal json for visit '%s': %v", id, err)
			logrus.Error(msg)
			return nil, errors.Wrap(err, msg)
		}
		visitors = append(visitors, value)
	}
	return visitors, nil
}

// IncreaseVisitCounter is a no-op and returns nil for all values.
//
// This function is unnecessary for the redis backend: we already
// have a redis LIST of visitors, and we can derive the visit count
// by calling redis.client.LLen(list) (which is a constant-time op)
// during GetEntryByID().  If we want the timestamp of the most recent
// visit we can pull the most recent visit off with redis.client.LIndex(0)
// (also constant-time) and reading the timetamp field.
func (r *Store) IncreaseVisitCounter(id string) error {
	return nil
}

// Close closes the connection to redis.
func (r *Store) Close() error {
	err := r.c.Close()
	if err != nil {
		msg := "Cloud not close the redis connection"
		logrus.Error(msg)
		return errors.Wrap(err, msg)
	}
	return err
}
