# Golang URL Shortener (Work in Progress)

[![Build Status](https://travis-ci.org/maxibanki/golang-url-shortener.svg?branch=master)](https://travis-ci.org/maxibanki/golang-url-shortener)
[![GoDoc](https://godoc.org/github.com/maxibanki/golang-url-shortener?status.svg)](https://godoc.org/github.com/maxibanki/golang-url-shortener)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxibanki/golang-url-shortener)](https://goreportcard.com/report/github.com/maxibanki/golang-url-shortener)
[![Coverage Status](https://coveralls.io/repos/github/maxibanki/golang-url-shortener/badge.svg?branch=master)](https://coveralls.io/github/maxibanki/golang-url-shortener?branch=master)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Download](https://api.bintray.com/packages/maxibanki/golang-url-shortener/travis-ci/images/download.svg?version=0.1) ](https://bintray.com/maxibanki/golang-url-shortener/travis-ci/0.1/link)
[![Docker Pulls](https://img.shields.io/docker/pulls/maxibanki/golang_url_shortener.svg)](https://hub.docker.com/r/maxibanki/golang_url_shortener/)

## Main Features

- URL Shortening
- Visitor Counting
- Expirable Links
- URL deletion
- Authorization System via OAuth 2.0 (Google, GitHub and Microsoft)
- High performance database with [bolt](https://github.com/boltdb/bolt)
- Easy [ShareX](https://github.com/ShareX/ShareX) integration
- Dockerizable

## [Webinterface](https://s.b0n.pl)

![Short URLs](https://user-images.githubusercontent.com/17984549/32700384-955d9336-c7c4-11e7-9fab-4141a86a375c.png)

---

![Generate ShareX Configuration](https://user-images.githubusercontent.com/17984549/32700395-cf9f057a-c7c4-11e7-9d2b-7523c8a95a20.png)

## Documenation

- [Installation](https://github.com/maxibanki/golang-url-shortener/wiki/Installation)
- [Configuration](https://github.com/maxibanki/golang-url-shortener/wiki/Configuration)
- [Setting up OAuth](https://github.com/maxibanki/golang-url-shortener/wiki/Setting-up-OAuth)
- [ShareX Usage](https://github.com/maxibanki/golang-url-shortener/wiki/ShareX)

## Why did you built this

Only because I just want to extend my current self hosted URL shorter (which was really messy code) with some more features and learn about new techniques like:

- Golang unit testing
- React
- Makefiles
- Travis CI
- Key / Value databases
- Dockerfile and Docker Image Creation
