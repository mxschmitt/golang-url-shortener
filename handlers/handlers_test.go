package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
)

const (
	testingDBName = "main.db"
	testURL       = "https://www.google.de/"
)

var server *httptest.Server

func TestCreateEntry(t *testing.T) {
	tt := []struct {
		name           string
		ignoreResponse bool
		contentType    string
		response       gin.H
		requestBody    URLUtil
		statusCode     int
	}{
		{
			name:           "body is nil",
			response:       gin.H{"error": "EOF"},
			statusCode:     http.StatusBadRequest,
			contentType:    "application/json; charset=utf-8",
			ignoreResponse: true,
		},
		{
			name: "short URL generation",
			requestBody: URLUtil{
				URL: "https://www.google.de/",
			},
			statusCode:  http.StatusOK,
			contentType: "application/json; charset=utf-8",
		},
		{
			name: "no valid URL",
			requestBody: URLUtil{
				URL: "this is really not a URL",
			},
			statusCode:     http.StatusBadRequest,
			contentType:    "application/json; charset=utf-8",
			response:       gin.H{"error": store.ErrNoValidURL.Error()},
			ignoreResponse: true,
		},
	}
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			if tc.requestBody.URL != "" {
				json, err := json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("could not marshal json: %v", err)
				}
				reqBody = json
			} else {
				reqBody = nil
			}
			respBody := createEntryWithJSON(t, reqBody, tc.contentType, tc.statusCode)
			if len(tc.response) > 0 {
				raw := makeJSON(t, tc.response)
				if string(respBody) != raw {
					t.Fatalf("expected body: %s; got: %s", tc.response, respBody)
				}
			}
			if tc.ignoreResponse {
				return
			}
			var parsed URLUtil
			err = json.Unmarshal(respBody, &parsed)
			if err != nil {
				t.Fatalf("could not unmarshal data: %v", err)
			}
			t.Run("test if shorted URL is correct", func(t *testing.T) {
				testRedirect(t, parsed.URL, tc.requestBody.URL)
			})
		})
	}
}

func TestHandleInfo(t *testing.T) {
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()

	t.Run("check existing entry", func(t *testing.T) {
		reqBody, err := json.Marshal(store.Entry{
			URL: testURL,
		})
		if err != nil {
			t.Fatalf("could not marshal json: %v", err)
		}
		respBody := createEntryWithJSON(t, reqBody, "application/json; charset=utf-8", http.StatusOK)
		var parsed URLUtil
		err = json.Unmarshal(respBody, &parsed)
		if err != nil {
			t.Fatalf("could not unmarshal data: %v", err)
		}
		body, err := json.Marshal(struct {
			ID string
		}{
			ID: strings.Replace(parsed.URL, server.URL+"/", "", 1),
		})
		if err != nil {
			t.Fatalf("could not marshal the body: %v", err)
		}
		resp, err := http.Post(server.URL+"/api/v1/info", "application/json; charset=utf-8", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		var entry store.Entry
		err = json.NewDecoder(resp.Body).Decode(&entry)
		if err != nil {
			t.Fatalf("could not unmarshal data: %v", err)
		}
		if entry.URL != testURL {
			t.Fatalf("url is not the expected one: %s; got: %s", testURL, entry.URL)
		}
	})
	t.Run("invalid body", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/info", "appplication/json", bytes.NewBuffer(nil))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		body = bytes.TrimSpace(body)
		raw := makeJSON(t, gin.H{
			"error": "Key: '.ID' Error:Field validation for 'ID' failed on the 'required' tag",
		})
		if string(body) != raw {
			t.Fatalf("body is not the expected one: %s", body)
		}
	})
	t.Run("no ID provided", func(t *testing.T) {
		if err != nil {
			t.Fatalf("could not marshal the body: %v", err)
		}
		resp, err := http.Post(server.URL+"/api/v1/info", "appplication/json", bytes.NewBufferString("{}"))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		raw := makeJSON(t, gin.H{
			"error": "Key: '.ID' Error:Field validation for 'ID' failed on the 'required' tag",
		})
		if string(body) != raw {
			t.Fatalf("body is not the expected one: %s", body)
		}
	})
}

func makeJSON(t *testing.T, data interface{}) string {
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("could not marshal json: %v", err)
	}
	return string(raw)
}

func createEntryWithJSON(t *testing.T, reqBody []byte, contentType string, statusCode int) []byte {
	resp, err := http.Post(server.URL+"/api/v1/create", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("could not post to backend %v", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read body: %v", err)
	}
	if resp.Header.Get("Content-Type") != contentType {
		t.Fatalf("content-type is not the expected one: %s; got: %s", contentType, resp.Header.Get("Content-Type"))
	}
	if resp.StatusCode != statusCode {
		t.Errorf("expected status %d; got %d", statusCode, resp.StatusCode)
	}
	return bytes.TrimSpace(respBody)
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
	resp, err := client.Do(&http.Request{
		URL: u,
	})
	if err != nil {
		t.Fatalf("could not do http request to shorted URL: %v", err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status code: %d; got: %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}
	if resp.Header.Get("Location") != longURL {
		t.Fatalf("redirect URL is not correct")
	}
}

func getBackend() (func(), error) {
	store, err := store.New(config.Store{
		DBPath:          testingDBName,
		ShortedIDLength: 4,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler, err := New(config.Handlers{
		ListenAddr: ":8080",
		Secret:     []byte(""),
		BaseURL:    "http://127.0.0.1",
	}, *store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create handler")
	}

	server = httptest.NewServer(handler.engine)
	return func() {
		server.Close()
		handler.CloseStore()
		os.Remove(testingDBName)
	}, nil
}
