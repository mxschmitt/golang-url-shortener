package store

import (
	"strings"
	"testing"

	"github.com/maxibanki/golang-url-shortener/util"
)

func TestGenerateRandomString(t *testing.T) {
	util.SetConfig(util.Configuration{
		DataDir:         "./data",
		ShortedIDLength: 4,
	})
	tt := []struct {
		name   string
		length int
	}{
		{"fourtytwo long", 42},
		{"sixteen long", 16},
		{"eighteen long", 19},
		{"zero long", 0},
		{"onehundretseventyfive long", 157},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rnd, err := generateRandomString(tc.length)
			if err != nil {
				t.Fatalf("could not generate random string: %v", err)
			}
			if len(rnd) != int(tc.length) {
				t.Fatalf("length of %s random string is %d not the expected one: %d", tc.name, len(rnd), tc.length)
			}
		})
	}
}

func TestNewStore(t *testing.T) {
	t.Run("create store with correct arguments", func(r *testing.T) {
		if err := util.ReadInConfig(); err != nil {
			t.Fatalf("could not read in config: %v", err)
		}
		store, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cleanup(store)
	})
}

func TestCreateEntry(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup(store)
	_, _, err = store.CreateEntry(Entry{}, "")
	if err != ErrNoValidURL {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i <= 100; i++ {
		_, _, err := store.CreateEntry(Entry{
			Public: EntryPublicData{
				URL: "https://golang.org/",
			},
		}, "")
		if err != nil && err != ErrGeneratingIDFailed {
			t.Fatalf("unexpected error during creating entry: %v", err)
		}
	}
}

func TestGetEntryByID(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup(store)
	_, err = store.GetEntryByID("something that not exists")
	if err != ErrNoEntryFound {
		t.Fatalf("could not get expected '%v' error: %v", ErrNoEntryFound, err)
	}
	_, err = store.GetEntryByID("")
	if err != ErrNoEntryFound {
		t.Fatalf("could not get expected '%v' error: %v", ErrNoEntryFound, err)
	}
}

func TestIncreaseVisitCounter(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	defer cleanup(store)
	id, _, err := store.CreateEntry(Entry{
		Public: EntryPublicData{
			URL: "https://golang.org/",
		},
	}, "")
	if err != nil {
		t.Fatalf("could not create entry: %v", err)
	}
	entryBeforeInc, err := store.GetEntryByID(id)
	if err != nil {
		t.Fatalf("could not get entry by id: %v", err)
	}
	if err = store.IncreaseVisitCounter(id); err != nil {
		t.Fatalf("could not increase visit counter %v", err)
	}
	entryAfterInc, err := store.GetEntryByID(id)
	if err != nil {
		t.Fatalf("could not get entry by id: %v", err)
	}
	if entryBeforeInc.Public.VisitCount+1 != entryAfterInc.Public.VisitCount {
		t.Fatalf("the increasement was not successful, the visit count is not correct")
	}
	if err = store.IncreaseVisitCounter(""); !strings.Contains(err.Error(), ErrNoEntryFound.Error()) {
		t.Fatalf("could not get expected '%v'; got: %v", ErrNoEntryFound, err)
	}
}

func TestDelete(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	defer cleanup(store)
	entryID, delHMac, err := store.CreateEntry(Entry{
		Public: EntryPublicData{
			URL: "https://golang.org/",
		},
	}, "")
	if err != nil {
		t.Fatalf("could not create entry: %v", err)
	}
	if err := store.DeleteEntry(entryID, delHMac); err != nil {
		t.Fatalf("could not delete entry: %v", err)
	}
	if _, err := store.GetEntryByID(entryID); err != ErrNoEntryFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetURLAndIncrease(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	defer cleanup(store)
	const url = "https://golang.org/"
	entryID, _, err := store.CreateEntry(Entry{
		Public: EntryPublicData{
			URL: url,
		},
	}, "")
	if err != nil {
		t.Fatalf("could not create entry: %v", err)
	}
	entryOne, err := store.GetEntryByID(entryID)
	if err != nil {
		t.Fatalf("could not get entry: %v", err)
	}
	entryURL, err := store.GetURLAndIncrease(entryID)
	if err != nil {
		t.Fatalf("could not get URL and increase the visitor counter: %v", err)
	}
	if entryURL != url {
		t.Fatalf("url is not the expected one")
	}
	entryTwo, err := store.GetEntryByID(entryID)
	if err != nil {
		t.Fatalf("could not get entry: %v", err)
	}
	if entryOne.Public.VisitCount+1 != entryTwo.Public.VisitCount {
		t.Fatalf("visitor count does not increase")
	}
}

func cleanup(s *Store) {
	s.Close()
}
