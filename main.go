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
	config, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "could not get config")
	}
	store, err := store.New(config.Store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler := handlers.New(config.Handlers, *store)
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
