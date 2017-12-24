package boltdb

import (
	"os"
	"testing"
	"time"

	"github.com/maxibanki/golang-url-shortener/stores/shared"
)

func getStore(t *testing.T) (*BoltStore, func()) {
	store, err := New("test.db")
	if err != nil {
		t.Errorf("could not get store: %v", err)
	}
	return store, func() {
		store.Close()
		os.Remove("test.db")
	}
}

func TestBoltDB(t *testing.T) {
	store, cleanup := getStore(t)
	givenEntryID := "x1df"
	givenEntry := shared.Entry{
		DeletionURL: "foo",
		RemoteAddr:  "127.0.0.1",
		Public: shared.EntryPublicData{
			CreatedOn: time.Now(),
			URL:       "google.com",
		},
	}
	if err := store.CreateEntry(givenEntry, givenEntryID, "google01234"); err != nil {
		t.Errorf("could not create entry: %v", err)
	}
	entryBeforeIncreasement, err := store.GetEntryByID(givenEntryID)
	if err != nil {
		t.Errorf("could not get entry: %v", err)
	}
	if err := store.IncreaseVisitCounter(givenEntryID); err != nil {
		t.Errorf("could not increase visit counter: %v", err)
	}
	entryAfterIncreasement, err := store.GetEntryByID(givenEntryID)
	if err != nil {
		t.Errorf("could not get entry: %v", err)
	}
	if entryBeforeIncreasement.Public.VisitCount+1 != entryAfterIncreasement.Public.VisitCount {
		t.Errorf("Visit counter hasn't increased; before: %d, after: %d", entryBeforeIncreasement.Public.VisitCount, entryAfterIncreasement.Public.VisitCount)
	}
	if err := store.RegisterVisitor(givenEntryID, "whooop", shared.Visitor{
		IP:      "foo",
		Referer: "foo",
	}); err != nil {
		t.Errorf("Failed to register visitor: %v", err)
	}
	visitors, err := store.GetVisitors(givenEntryID)
	if err != nil {
		t.Errorf("could not get visitors: %v", err)
	}
	if len(visitors) != 1 {
		t.Errorf("Expected visitor length: %d; got: %d", err)
	}
	if err := store.DeleteEntry(givenEntryID); err != nil {
		t.Errorf("could not delte entry: %v", err)
	}
	cleanup()
}
