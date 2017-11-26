package main

import (
	"os"
	"os/signal"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"

	"github.com/maxibanki/golang-url-shortener/handlers"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/pkg/errors"
)

func main() {
	os.Setenv("GUS_SHORTED_ID_LENGTH", "4")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	close, err := initShortener()
	if err != nil {
		logrus.Fatalf("could not init shortener: %v", err)
	}
	<-stop
	logrus.Println("Shutting down...")
	close()
}

func initShortener() (func(), error) {
	if err := util.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "could not reload config file")
	}
	if util.GetConfig().EnableDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
	store, err := store.New()
	if err != nil {
		return nil, errors.Wrap(err, "could not create store")
	}
	handler, err := handlers.New(*store)
	if err != nil {
		return nil, errors.Wrap(err, "could not create handlers")
	}
	go func() {
		if err := handler.Listen(); err != nil {
			logrus.Fatalf("could not listen to http handlers: %v", err)
		}
	}()
	return func() {
		if err = handler.CloseStore(); err != nil {
			logrus.Printf("failed to stop the handlers: %v", err)
		}
	}, nil
}
