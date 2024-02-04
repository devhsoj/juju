package main

import (
	"juju"
	"log"
	"os"
)

func main() {
	server := juju.Server{}

	var listenAddress = "localhost:9261"

	if len(os.Getenv("LISTEN_ADDRESS")) > 0 {
		listenAddress = os.Getenv("LISTEN_ADDRESS")
	} else if len(os.Getenv("HOSTNAME")) > 0 && os.Getenv("DOCKER") == "true" {
		listenAddress = os.Getenv("HOSTNAME") + ":9261"
	}

	if err := server.Start(listenAddress); err != nil {
		log.Printf("failed to start juju server: %s", err)
	}
}
