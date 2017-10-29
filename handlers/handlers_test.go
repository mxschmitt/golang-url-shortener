package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/maxibanki/golang-url-shorter/store"
	"github.com/pkg/errors"
)

const (
	baseURL       = "http://myshorter"
	testingDBName = "main.db"
)

var server *httptest.Server

func TestCreateEntryJSON(t *testing.T) {
	tt := []struct {
		name           string
		ignoreResponse bool
		contentType    string
		response       string
		responseBody   URLUtil
		statusCode     int
	}{
		{
			name:           "body is nil",
			response:       "invalid request, body is nil",
			statusCode:     http.StatusBadRequest,
			contentType:    "appication/json",
			ignoreResponse: true,
		},
		{
			name: "short URL generation",
			responseBody: URLUtil{
				URL: "https://www.google.de/",
			},
			statusCode:  http.StatusOK,
			contentType: "appication/json",
		},
	}
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// build body for the create URL http request
			var reqBody *bytes.Buffer
			if tc.responseBody.URL != "" {
				json, err := json.Marshal(tc.responseBody)
				if err != nil {
					t.Fatalf("could not marshal json: %v", err)
				}
				reqBody = bytes.NewBuffer(json)
			} else {
				reqBody = bytes.NewBuffer(nil)
			}
			resp, err := http.Post(server.URL+"/api/v1/create", "application/json", reqBody)
			if err != nil {
				t.Fatalf("could not create post request: %v", err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}
			body = bytes.TrimSpace(body)
			if tc.ignoreResponse {
				return
			}
			if resp.StatusCode != tc.statusCode {
				t.Errorf("expected status %d; got %d", tc.statusCode, resp.StatusCode)
			}
			if tc.response != "" {
				if string(body) != string(tc.response) {
					t.Fatalf("expected body: %s; got: %s", tc.response, body)
				}
			}
			var parsed URLUtil
			err = json.Unmarshal(body, &parsed)
			if err != nil {
				t.Fatalf("could not unmarshal data: %v", err)
			}
			t.Run("test if shorted URL is correct", func(t *testing.T) {
				testRedirect(t, parsed.URL, tc.responseBody.URL)
			})
		})
	}
}

func testRedirect(t *testing.T, shortURL, longURL string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}, // don't follow redirects
	}
	u, err := url.Parse(shortURL)
	if err != nil {
		t.Fatalf("could not parse shorted URL: %v", err)
	}
	respShort, err := client.Do(&http.Request{
		URL: u,
	})
	if err != nil {
		t.Fatalf("could not do http request to shorted URL: %v", err)
	}
	if respShort.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status code: %d; got: %d", http.StatusTemporaryRedirect, respShort.StatusCode)
	}
	if respShort.Header.Get("Location") != longURL {
		t.Fatalf("redirect URL is not correct")
	}
}

func getBackend() (func(), error) {
	store, err := store.New(testingDBName, 4)
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler := New(":8080", *store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create handler")
	}
	server = httptest.NewServer(handler.handlers())
	return func() {
		server.Close()
		handler.Stop()
		os.Remove(testingDBName)
	}, nil
}
