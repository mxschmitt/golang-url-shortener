package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/maxibanki/golang-url-shortener/handlers"
	"github.com/maxibanki/golang-url-shortener/store"
)

func main() {
	dbPath := "main.db"
	listenAddr := ":8080"
	idLength := uint(4)
	if os.Getenv("SHORTENER_DB_PATH") != "" {
		dbPath = os.Getenv("SHORTENER_DB_PATH")
	}
	if os.Getenv("SHORTENER_LISTEN_ADDR") != "" {
		listenAddr = os.Getenv("SHORTENER_LISTEN_ADDR")
	}
	if os.Getenv("SHORTENER_ID_LENGTH") != "" {
		length, err := strconv.ParseUint(os.Getenv("SHORTENER_ID_LENGTH"), 10, 32)
		if err != nil {
			log.Fatalf("could not convert ID length into an uint: %v", err)
		}
		idLength = uint(length)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	store, err := store.New(dbPath, idLength)
	if err != nil {
		log.Fatalf("could not create store: %v", err)
	}
	handler := handlers.New(listenAddr, *store)
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
