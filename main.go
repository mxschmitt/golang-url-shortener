package main

import (
	"os"
	"os/signal"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"

	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/handlers"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
	log.Out = ansicolor.NewAnsiColorWriter(os.Stdout)
	close, err := initShortener(log)
	if err != nil {
		log.Fatalf("could not init shortener: %v", err)
	}
	<-stop
	log.Println("Shutting down...")
	close()
}

func initShortener(log *logrus.Logger) (func(), error) {
	if err := config.Preload(); err != nil {
		return nil, errors.Wrap(err, "could not get config")
	}
	conf := config.Get()
	store, err := store.New(conf.Store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler, err := handlers.New(conf.Handlers, *store, log, false)
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
