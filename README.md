# Golang URL Shortener using BoltDB

[![Build Status](https://travis-ci.org/maxibanki/golang-url-shortener.svg?branch=master)](https://travis-ci.org/maxibanki/golang-url-shortener)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxibanki/golang-url-shortener)](https://goreportcard.com/report/github.com/maxibanki/golang-url-shortener)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
## Features:

- URL Shortening with visitor counting
- Deletion URLs
- Authorization System
- High Performance database with [bolt](https://github.com/boltdb/bolt)
- ShareX integration
- Easy Docker Deployment

## Server Installation

### Standard

```bash
git clone https://github.com/maxibanki/golang-url-shortener
go get -v ./...
go build
./golang-url-shortener
```
### Docker Compose

- Only execute the [docker-compose.yml](docker-compose.yml) and adjust the enviroment variables to your needs.

### Envirment Variables:

| Envirment Variable | Description | Default Value |
| ------------------ | ----------- | ------------- |
| SHORTENER_DB_PATH  | Relative or absolute path to the bolt DB | main.db |
| SHORTENER_LISTEN_ADDR | Adress to which the http server should listen to | :8080 |
| SHORTENER_ID_LENGTH | Length of the random short URL id | 4 |

## Clients:

### [ShareX](https://github.com/ShareX/ShareX) Configuration

This URL Shortener has fully support with ShareX. To use it, just import the configuration to your ShareX. For that you need to open the `Destination settings` => `Other / Custom uploaders` => `Import` => `From Clipboard`.

After you've done this, you need to set it as your standard URL Shortener. For that go back into your main menu => `Destinations` => `URL Shortener` => `Custom URL Shortener`.

```json
{
  "Name": "Golang URL Shortener",
  "DestinationType": "URLShortener",
  "RequestType": "POST",
  "RequestURL": "http://127.0.0.1:8080/api/v1/create",
  "Arguments": {
    "URL": "$input$"
  },
  "ResponseType": "Text",
  "URL": "$json:URL$"
}
```

### Curl

## TODO

- Authentification
- Deletion
- Github publishing
- Add shields:
  - downloads
  - travis
  - godoc
  - license