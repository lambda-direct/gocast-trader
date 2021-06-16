package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/lambda-direct/gocast-trader/fetcher"
	"github.com/lambda-direct/gocast-trader/persister"
)

func main() {
	errc := make(chan error)

	fetchClient := fetcher.New()
	go fetchClient.Fetch()

	persisterClient := persister.New(fetchClient)
	go persisterClient.Watch(errc)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	select {
	case err := <-errc:
		panic(err)

	case sig := <-signalChan:
		log.Printf("Received signal %s", sig)
	}
}
