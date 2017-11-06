package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/urakozz/go-json-schema-generator"
)

func main() {
	schemaPath := flag.String("path", "schema.json", "location of the converted schema")
	flag.Parse()
	schema := generator.Generate(&config.Configuration{})
	if err := ioutil.WriteFile(*schemaPath, []byte(schema), 644); err != nil {
		log.Fatalf("could not write schema: %v", err)
	}
}
