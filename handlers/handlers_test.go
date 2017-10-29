package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/maxibanki/golang-url-shorter/store"
	"github.com/pkg/errors"
)

const (
	baseURL       = "http://myshorter"
	testingDBName = "main.db"
	testURL       = "https://www.google.de/"
)

var server *httptest.Server

func TestCreateEntryJSON(t *testing.T) {
	tt := []struct {
		name           string
		ignoreResponse bool
		contentType    string
		response       string
		requestBody    URLUtil
		statusCode     int
	}{
		{
			name:           "body is nil",
			response:       "could not decode JSON: EOF",
			statusCode:     http.StatusBadRequest,
			contentType:    "appication/json",
			ignoreResponse: true,
		},
		{
			name: "short URL generation",
			requestBody: URLUtil{
				URL: "https://www.google.de/",
			},
			statusCode:  http.StatusOK,
			contentType: "appication/json",
		},
		{
			name: "no valid URL",
			requestBody: URLUtil{
				URL: "this is really not a URL",
			},
			statusCode:     http.StatusBadRequest,
			contentType:    "appication/json",
			response:       store.ErrNoValidURL.Error(),
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
			// build body for the create URL http request
			var reqBody *bytes.Buffer
			if tc.requestBody.URL != "" {
				json, err := json.Marshal(tc.requestBody)
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
			if resp.StatusCode != tc.statusCode {
				t.Errorf("expected status %d; got %d", tc.statusCode, resp.StatusCode)
			}
			if tc.response != "" {
				if string(body) != string(tc.response) {
					t.Fatalf("expected body: %s; got: %s", tc.response, body)
				}
			}
			if tc.ignoreResponse {
				return
			}
			var parsed URLUtil
			err = json.Unmarshal(body, &parsed)
			if err != nil {
				t.Fatalf("could not unmarshal data: %v", err)
			}
			t.Run("test if shorted URL is correct", func(t *testing.T) {
				testRedirect(t, parsed.URL, tc.requestBody.URL)
			})
		})
	}
}

func TestCreateEntryMultipart(t *testing.T) {
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()

	t.Run("valid request", func(t *testing.T) {
		// Prepare a form that you will submit to that URL.
		var b bytes.Buffer
		multipartWriter := multipart.NewWriter(&b)
		formWriter, err := multipartWriter.CreateFormField("URL")
		if err != nil {
			t.Fatalf("could not create form field: %v", err)
		}
		formWriter.Write([]byte(testURL))
		multipartWriter.Close()

		resp, err := http.Post(server.URL+"/api/v1/create", multipartWriter.FormDataContentType(), &b)
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		var parsed URLUtil
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		if err != nil {
			t.Fatalf("could not unmarshal data: %v", err)
		}
		t.Run("test if shorted URL is correct", func(t *testing.T) {
			testRedirect(t, parsed.URL, testURL)
		})
	})

	t.Run("invalid url", func(t *testing.T) {
		// Prepare a form that you will submit to that URL.
		var b bytes.Buffer
		multipartWriter := multipart.NewWriter(&b)
		formWriter, err := multipartWriter.CreateFormField("URL")
		if err != nil {
			t.Fatalf("could not create form field: %v", err)
		}
		formWriter.Write([]byte("this is definitely not a valid url"))
		multipartWriter.Close()

		resp, err := http.Post(server.URL+"/api/v1/create", multipartWriter.FormDataContentType(), &b)
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		if string(body) != store.ErrNoValidURL.Error() {
			t.Fatalf("received unexpected response: %s", body)
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		// Prepare a form that you will submit to that URL.
		var b bytes.Buffer
		multipartWriter := multipart.NewWriter(&b)
		multipartWriter.Close()

		resp, err := http.Post(server.URL+"/api/v1/create", multipartWriter.FormDataContentType(), &b)
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body")
		}
		if string(body) != "URL key does not exist" {
			t.Fatalf("body has not the excepted payload; got: %s", body)
		}
	})
}

func TestCreateEntryForm(t *testing.T) {
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()

	t.Run("valid request", func(t *testing.T) {
		data := url.Values{}
		data.Set("URL", testURL)

		resp, err := http.Post(server.URL+"/api/v1/create", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		var parsed URLUtil
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		if err != nil {
			t.Fatalf("could not unmarshal data: %v", err)
		}
		t.Run("test if shorted URL is correct", func(t *testing.T) {
			testRedirect(t, parsed.URL, testURL)
		})
	})

	t.Run("invalid request", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/v1/create", "application/x-www-form-urlencoded", bytes.NewBufferString(url.Values{}.Encode()))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		if string(body) != "URL key does not exist" {
			t.Fatalf("received unexpected response: %s", body)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		data := url.Values{}
		data.Set("URL", "this is definitely not a valid url")

		resp, err := http.Post(server.URL+"/api/v1/create", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
		if err != nil {
			t.Fatalf("could not post to the backend: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		if string(body) != store.ErrNoValidURL.Error() {
			t.Fatalf("received unexpected response: %s", body)
		}
	})
}

func TestHandleInfo(t *testing.T) {
	cleanup, err := getBackend()
	if err != nil {
		t.Fatalf("could not create backend: %v", err)
	}
	defer cleanup()

	t.Run("check existing entry", func(t *testing.T) {
		body, err := json.Marshal(store.Entry{
			URL: testURL,
		})
		if err != nil {
			t.Fatalf("could not marshal json: %v", err)
		}
		resp, err := http.Post(server.URL+"/api/v1/create", "application/json", bytes.NewBuffer(body))
		var parsed URLUtil
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		if err != nil {
			t.Fatalf("could not unmarshal data: %v", err)
		}
		id := strings.Replace(parsed.URL, server.URL+"/", "", 1)
		body, err = json.Marshal(struct {
			ID string
		}{
			ID: id,
		})
		if err != nil {
			t.Fatalf("could not marshal the body: %v", err)
		}
		resp, err = http.Post(server.URL+"/api/v1/info", "appplication/json", bytes.NewBuffer(body))
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
		body = bytes.TrimSpace(body)
		if err != nil {
			t.Fatalf("could not read the body: %v", err)
		}
		if string(body) != "could not decode JSON: EOF" {
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
		if string(body) != "no ID provided" {
			t.Fatalf("body is not the expected one: %s", body)
		}
	})
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
