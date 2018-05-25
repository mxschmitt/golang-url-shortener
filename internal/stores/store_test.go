package stores

import (
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/mxschmitt/golang-url-shortener/internal/stores/shared"

	"github.com/mxschmitt/golang-url-shortener/internal/util"
)

var testData = struct {
	ID            string
	Password      string
	oAuthProvider string
	oAuthID       string
	Entry         shared.Entry
	Visitor       shared.Visitor
	DataDir       string
}{
	ID:            "such-a-great-id",
	Password:      "sooo secret",
	oAuthProvider: "google",
	oAuthID:       "12345678",
	Entry: shared.Entry{
		Public: shared.EntryPublicData{
			URL: "https://google.com",
		},
		RemoteAddr: "203.0.113.6",
	},
	Visitor: shared.Visitor{
		IP:      "foo",
		Referer: "foo",
	},
	DataDir: "./data",
}

func TestGenerateRandomString(t *testing.T) {
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

func TestStore(t *testing.T) {
	util.SetConfig(util.Configuration{
		DataDir:         testData.DataDir,
		Backend:         "boltdb",
		ShortedIDLength: 4,
	})
	if err := os.MkdirAll(testData.DataDir, 0755); err != nil {
		t.Errorf("could not create data dir: %v", err)
	}
	store, err := New()
	if err != nil {
		t.Errorf("could not create store: %v", err)
	}
	entryID, deletionHmac, err := store.CreateEntry(testData.Entry, "", testData.Password)
	if err != nil {
		t.Errorf("could not create entry: %v", err)
	}
	entryBeforeIncreasement, err := store.GetEntryByID(entryID)
	if err != nil {
		t.Errorf("could not get entry: %v", err)
	}
	entry, err := store.GetEntryAndIncrease(entryID)
	if err != nil {
		t.Errorf("could not increase entry: %v", err)
	}
	entryAfterIncreasement, err := store.GetEntryByID(entryID)
	if err != nil {
		t.Errorf("could not get entry: %v", err)
	}
	if entryBeforeIncreasement.Public.VisitCount+1 != entryAfterIncreasement.Public.VisitCount {
		t.Errorf("Visit counter hasn't increased; before: %d, after: %d", entryBeforeIncreasement.Public.VisitCount, entryAfterIncreasement.Public.VisitCount)
	}
	if entryAfterIncreasement.Public.VisitCount != entry.Public.VisitCount {
		t.Errorf("returned entry from increasement does not mach visitor count; got: %d; expected: %d", entry.Public.VisitCount, entryAfterIncreasement.Public.VisitCount)
	}
	store.RegisterVisit(entryID, testData.Visitor)
	visitors, err := store.GetVisitors(entryID)
	if err != nil {
		t.Errorf("could not get visitors: %v", err)
	}
	visitor := visitors[0]
	if visitor.IP != testData.Visitor.IP && visitor.Referer != testData.Visitor.Referer {
		t.Errorf("received visitor does not match")
	}
	if err := store.DeleteEntry(entryID, deletionHmac); err != nil {
		t.Errorf("could not delete entry: %v", err)
	}
	if _, err := store.GetEntryByID(entryID); errors.Cause(err) != shared.ErrNoEntryFound {
		t.Errorf("error is not expected one: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Errorf("could not close store: %v", err)
	}
	if err := os.RemoveAll(testData.DataDir); err != nil {
		t.Errorf("could not remove database: %v", err)
	}
}
