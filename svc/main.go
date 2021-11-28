package main

import (
	"context"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")

	if serviceName == "" {
		log.Fatal("SERVICE_NAME env var must be set")
	}

	log.Info("setting up dependencies")

	// Setup dependencies
	deps, err := setupDependencies(serviceName)
	if err != nil {
		log.Fatalf("unable to setup deps: %s", err)
	}

	log.Info("starting consumer")

	// Run consumer
	consumer := &Consumer{
		ServiceName: serviceName,
		Deps:        deps,
	}

	go deps.EventBus.Consume(context.Background(), nil, consumer.ConsumeFunc)

	log.Infof("'%s' service started", serviceName)

	// Run forever
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
