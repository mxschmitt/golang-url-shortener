package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/maxibanki/golang-url-shorter/handlers"
	"github.com/maxibanki/golang-url-shorter/store"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	store, err := store.New("main.db", 4)
	if err != nil {
		log.Fatalf("could not create store: %v", err)
	}
	handler := handlers.New(":8080", *store)
	go func() {
		err := handler.Listen()
		if err != nil {
			log.Fatalf("could not listen to http handlers: %v", err)
		}
	}()
	<-stop
	log.Println("Shutting down...")
	err = handler.Stop()
	if err != nil {
		log.Printf("failed to stop the handlers: %v", err)
	}
}
