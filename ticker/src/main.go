package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/lambda-direct/gocast-trader/common/src/env"
	"github.com/lambda-direct/gocast-trader/ticker/src/fetcher"
	"github.com/lambda-direct/gocast-trader/ticker/src/persister"
)

func main() {
	s := new(env.Spec)

	envconfig.MustProcess("", s)

	errc := make(chan error)

	fetchClient := fetcher.New()
	go fetchClient.Fetch()

	persisterClient := persister.New(fetchClient, s)
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
