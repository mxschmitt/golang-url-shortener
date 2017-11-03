package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/handlers"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
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
	if err := config.Preload(); err != nil {
		return nil, errors.Wrap(err, "could not get config")
	}
	conf := config.Get()
	store, err := store.New(conf.Store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler, err := handlers.New(conf.Handlers, *store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create handlers")
	}
	go func() {
		if err := handler.Listen(); err != nil {
			log.Fatalf("could not listen to http handlers: %v", err)
		}
	}()
	return func() {
		if err = handler.CloseStore(); err != nil {
			log.Printf("failed to stop the handlers: %v", err)
		}
	}, nil
}
