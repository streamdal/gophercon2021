package main

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("setting up dependencies")

	// Setup event bus
	deps, err := setupDependencies()
	if err != nil {
		log.Fatalf("unable to setup deps: %s", err)
	}

	log.Info("starting consumer")

	// Run consumer
	consumer := &Consumer{deps}

	go deps.EventBus.Consume(context.Background(), nil, consumer.ConsumeFunc)

	log.Info("order service started")

	// Run forever
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
