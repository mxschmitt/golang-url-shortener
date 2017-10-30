package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/maxibanki/golang-url-shortener/handlers"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	close, err := initShortener()
	if err != nil {
		log.Fatalf("could not init shortener: %v", err)
	}
	<-stop
	log.Println("Shutting down...")
	close()
}

func initShortener() (func(), error) {
	var config struct {
		DBPath          string `yaml:"DBPath"`
		ListenAddr      string `yaml:"ListenAddr"`
		ShortedIDLength int    `yaml:"ShortedIDLength"`
	}
	ex, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "could not get executable path")
	}
	file, err := ioutil.ReadFile(filepath.Join(filepath.Dir(ex), "config.yml"))
	if err != nil {
		return nil, errors.Wrap(err, "could not read configuration file: %v")
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal yaml file")
	}
	store, err := store.New(config.DBPath, config.ShortedIDLength)
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler := handlers.New(config.ListenAddr, *store)
	go func() {
		err := handler.Listen()
		if err != nil {
			log.Fatalf("could not listen to http handlers: %v", err)
		}
	}()
	return func() {
		err = handler.CloseStore()
		if err != nil {
			log.Printf("failed to stop the handlers: %v", err)
		}
	}, nil
}
