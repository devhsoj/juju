package main

import (
	"juju"
	"log"
	"os"
)

func main() {
	server := juju.Server{}

	if err := server.Start(os.Getenv("LISTEN_ADDRESS")); err != nil {
		log.Printf("failed to start juju server: %s", err)
	}
}
