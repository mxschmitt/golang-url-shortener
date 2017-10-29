# Golang URL Shortener using BoltDB

## Features:

- URL Shortening with visiter counting
- Delete an entry
- Authorization
- Storing using BoltDB
- Easy ShareX integration
- Selfhosted

## Installation

### Standard

```bash
go get -v ./...
go run -v main.go
```
### Docker Compose

Only execute the [docker-compose.yml](docker-compose.yml) and adjust the enviroment variables to your needs.

## [ShareX](https://github.com/ShareX/ShareX) Configuration

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



## TODOs

- github publishing
- authentification
- deletion
  - ShareX example