package store

import (
	"os"
	"testing"
)

const (
	testingDBName = "test.db"
)

func TestGenerateRandomString(t *testing.T) {
	tt := []struct {
		name   string
		length uint
	}{
		{"fourtytwo long", 42},
		{"sixteen long", 16},
		{"eighteen long", 19},
		{"zero long", 0},
		{"onehundretseventyfive long", 157},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rnd := generateRandomString(tc.length)
			if len(rnd) != int(tc.length) {
				t.Fatalf("length of %s random string is %d not the expected one: %d", tc.name, len(rnd), tc.length)
			}
		})
	}
}

func TestNewStore(t *testing.T) {
	t.Run("create store without file name provided", func(r *testing.T) {
		_, err := New("", 4)
		if err.Error() != "could not open bolt DB database: open : The system cannot find the file specified." {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("create store with correct arguments", func(r *testing.T) {
		store, err := New(testingDBName, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cleanup(store)
	})
}

func TestCreateEntry(t *testing.T) {
	store, err := New(testingDBName, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup(store)
	_, err = store.CreateEntry("", "")
	if err != ErrNoValidURL {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i <= 100; i++ {
		_, err := store.CreateEntry("https://golang.org/", "")
		if err != nil && err != ErrGeneratingTriesFailed {
			t.Fatalf("unexpected error during creating entry: %v", err)
		}
	}
}

func TestGetEntryByID(t *testing.T) {
	store, err := New(testingDBName, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup(store)
	_, err = store.GetEntryByID("something that not exists")
	if err != ErrNoEntryFound {
		t.Fatalf("could not get expected '%v' error: %v", ErrNoEntryFound, err)
	}
	_, err = store.GetEntryByID("")
	if err != ErrIDIsEmpty {
		t.Fatalf("could not get expected '%v' error: %v", ErrIDIsEmpty, err)
	}
}

func TestIncreaseVisitCounter(t *testing.T) {
	store, err := New(testingDBName, 4)
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	defer cleanup(store)
	id, err := store.CreateEntry("https://golang.org/", "")
	if err != nil {
		t.Fatalf("could not create entry: %v", err)
	}
	entryBeforeInc, err := store.GetEntryByID(id)
	if err != nil {
		t.Fatalf("could not get entry by id: %v", err)
	}
	err = store.IncreaseVisitCounter(id)
	if err != nil {
		t.Fatalf("could not increase visit counter %v", err)
	}
	entryAfterInc, err := store.GetEntryByID(id)
	if err != nil {
		t.Fatalf("could not get entry by id: %v", err)
	}
	if entryBeforeInc.VisitCount+1 != entryAfterInc.VisitCount {
		t.Fatalf("the increasement was not successful, the visit count is not correct")
	}
	err = store.IncreaseVisitCounter("")
	if err != ErrIDIsEmpty {
		t.Fatalf("could not get expected '%v' error: %v", ErrIDIsEmpty, err)
	}
}

func cleanup(s *Store) {
	s.Close()
	os.Remove(testingDBName)
}
