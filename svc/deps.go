package main

import (
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/batchcorp/rabbit"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

type Dependencies struct {
	EventBus     rabbit.IRabbit
	EtcdClient   *clientv3.Client
	StatsDClient *statsd.Client
	State        *State
}

func setupDependencies(serviceName string) (*Dependencies, error) {
	deps := &Dependencies{}

	// Setup event bus
	eventBus, err := rabbit.New(&rabbit.Options{
		URLs:         []string{"amqp://localhost"},
		QueueName:    serviceName + "-queue",
		QueueDeclare: true,
		QueueDurable: true,
		AutoAck:      true,
		Bindings: []rabbit.Binding{
			{
				ExchangeName:    "events",
				ExchangeType:    "topic",
				ExchangeDeclare: true,
				ExchangeDurable: true,
				BindingKeys:     []string{"#"},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new rabbit backend")
	}

	deps.EventBus = eventBus

	statsdClient, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup statsd client")
	}

	deps.StatsDClient = statsdClient

	// Let's populate statsd with dummy metrics (so we can reference the metrics in graphs)
	deps.StatsDClient.Count(serviceName+"_new_order_ok", 0, nil, 1)
	deps.StatsDClient.Count(serviceName+"_new_order_skipped", 0, nil, 1)

	// Setup state keeping
	state, err := NewState(serviceName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup state")
	}

	deps.State = state

	return deps, nil
}
