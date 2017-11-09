# Golang URL Shortener (Work in Progress)

[![Build Status](https://travis-ci.org/maxibanki/golang-url-shortener.svg?branch=master)](https://travis-ci.org/maxibanki/golang-url-shortener)
[![GoDoc](https://godoc.org/github.com/maxibanki/golang-url-shortener?status.svg)](https://godoc.org/github.com/maxibanki/golang-url-shortener)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxibanki/golang-url-shortener)](https://goreportcard.com/report/github.com/maxibanki/golang-url-shortener)
[![Coverage Status](https://coveralls.io/repos/github/maxibanki/golang-url-shortener/badge.svg?branch=master)](https://coveralls.io/github/maxibanki/golang-url-shortener?branch=master)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Waffle.io - Columns and their card count](https://badge.waffle.io/maxibanki/golang-url-shortener.png?columns=all)](https://waffle.io/maxibanki/golang-url-shortener?utm_source=badge)
[![Download](https://api.bintray.com/packages/maxibanki/golang-url-shortener/travis-ci/images/download.svg?version=0.1) ](https://bintray.com/maxibanki/golang-url-shortener/travis-ci/0.1/link)

## Main Features

- URL Shortening
- Visitor Counting
- Expirable Links
- URL deletion
- Authorization System via OAuth 2.0 from Google (more providers following)
- High performance database with [bolt](https://github.com/boltdb/bolt)
- Easy [ShareX](https://github.com/ShareX/ShareX) integration
- Dockerizable

## Server Installation

### Standard

Download the package for your architecture and operating system from [bintray](https://bintray.com/maxibanki/golang-url-shortener/travis-ci) and extract it.

### Docker

TODO

## Configuration

The configuration is a JSON file, an example is located [here](build/config.json). If your editor supports intellisense by using a schema (e.g. [VS Code](https://github.com/Microsoft/vscode)) then you can simply press space for auto completion. The config parameters should be really self explaining, but [here](build/schema.md) is a detailed description for all of these:

## OAuth

### Google

Visit [console.cloud.google.com](https://console.cloud.google.com) and create or use an existing project, go to `APIs & Services` -> `Credentials` and create there an `OAuth Client-ID` for the application type `Webapplicaton`. There you get the Client-ID and the ClientSecret for your configuration. It's important, that you set in the Google Cloud Platform `YOUR_URL/api/v1/callback` as authorized redirect URL.

## Clients

### General

In general the `POST` endpoints can be called, by using one of the following techniques:

- application/json
- application/x-www-form-urlencoded
- multipart/form-data

For all the endpoints which are on `/api/v1/protected` there is the `Authorization` header required.

### [ShareX](https://github.com/ShareX/ShareX)

For ShareX usage, we refer to the menu item in the frontend where your configuration will be generated. There are further information for the detailed use.

## Why did you built this

Just only because I want to extend my current self hosted URL shorter with some features and learn about new techniques like:

- Golang unit testing
- React
- Makefiles
- Travis CI
- Key / Value databases

## Utils

### Update Config Documentation

```
yarn global add jsonschema-md
go run build/schema.go
jsonschema-md.cmd build/schema.json > build/schema.md
```

After that adjust the title to `Configuration` and the description to `Golang URL Shortener Configuration`.