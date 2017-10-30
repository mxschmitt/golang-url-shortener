# Golang URL Shortener

[![Build Status](https://travis-ci.org/maxibanki/golang-url-shortener.svg?branch=master)](https://travis-ci.org/maxibanki/golang-url-shortener)
[![GoDoc](https://godoc.org/github.com/maxibanki/golang-url-shortener?status.svg)](https://godoc.org/github.com/maxibanki/golang-url-shortener)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxibanki/golang-url-shortener)](https://goreportcard.com/report/github.com/maxibanki/golang-url-shortener)
[![Coverage Status](https://coveralls.io/repos/github/maxibanki/golang-url-shortener/badge.svg?branch=master)](https://coveralls.io/github/maxibanki/golang-url-shortener?branch=master)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
## Main Features:

- URL Shortening
- Visitor Counting
- URL deletion 
- Authorization System
- High performance database with [bolt](https://github.com/boltdb/bolt)
- [ShareX](https://github.com/ShareX/ShareX) integration
- Easy Docker Deployment

## Server Installation

### Standard

Since we don't provide prebuild binaries, you have to build it yourself. For that you need Golang and Git installed on your system.

```bash
git clone https://github.com/maxibanki/golang-url-shortener # Clone repository
cd golang-url-shortener                                     # Go into it
go get -v ./...                                             # Fetch dependencies
go build                                                    # Build executable
./golang-url-shortener                                      # Run it
```
### Docker Compose

Only execute the [docker-compose.yml](docker-compose.yml) and adjust the environment variables to your needs.

### Environment Variables:

| Environment Variable | Description | Default Value |
| ------------------ | ----------- | ------------- |
| SHORTENER_DB_PATH  | Relative or absolute path to the bolt DB | main.db |
| SHORTENER_LISTEN_ADDR | Address to which the http server should listen to | :8080 |
| SHORTENER_ID_LENGTH | Length of the random short URL id | 4 |

## Clients:

### [ShareX](https://github.com/ShareX/ShareX) Configuration

This URL Shortener has fully support with ShareX. To use it, just import the configuration to your ShareX. For that you need to open the `Destination settings` => `Other / Custom uploaders` => `Import` => `From Clipboard` menu.

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

#### Create

Request:
```bash
curl -X POST -H 'Content-Type: application/json' -d '{"URL":"https://www.google.de/"}' http://127.0.0.1:8080/api/v1/create
```
Response:
```json
{
    "URL": "http://127.0.0.1:8080/dgUV",
}
```

#### Info

Request:
```bash
$ curl -X POST -H 'Content-Type: application/json' -d '{"ID":"dgUV"}' http://127.0.0.1:8080/api/v1/info
```
Response:
```json
{
    "URL": "https://google.com/",
    "VisitCount": 1,
    "CreatedOn": "2017-10-29T23:35:48.2977548+01:00",
    "LastVisit": "2017-10-29T23:36:14.193236+01:00"
}
```

### HTTP Endpoints:

#### `/api/v1/create` POST

Create is the handler for creating entries, you need to provide only an URL. The response will always be JSON encoded and contain an URL with the short link.

There is a mechanism integrated, that you can call this endpoint with the following techniques:
- application/json
- application/x-www-form-urlencoded
- multipart/form-data

In all cases only add the long URL as a field with the key `URL` and you will get the response with the short URL.

####  `/api/v1/info` POST

This handler returns the information about an entry. This includes:
- Created At
- Last Visit
- Visitor counter

To use this, POST a JSON with the field `id` to the endpoint. It will return a JSON with the data.

## TODO

Next changes sorted by priority

- Update http stuff to the gin framework
- Authorization
- Deletion
- Test docker-compose installation