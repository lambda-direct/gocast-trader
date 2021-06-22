package main

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/lambda-direct/gocast-trader/common/src/env"
	"github.com/lambda-direct/gocast-trader/ticker/src/fetcher"
	"github.com/lambda-direct/gocast-trader/ticker/src/lock"
	"github.com/lambda-direct/gocast-trader/ticker/src/persister"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	s := new(env.Spec)

	envconfig.MustProcess("", s)

	errc := make(chan error)

	l, err := lock.New(s)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	fetchClient := fetcher.New()
	go fetchClient.Fetch()

	persisterClient := persister.New(ctx, fetchClient, s, l)
	go persisterClient.Watch(errc)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errc:
		panic(err)

	case sig := <-signalChan:
		log.Printf("Received signal %s", sig)
		cancel()
		if err := l.Release(); err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
	}
}
