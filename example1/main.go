package	main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/batchcorp/rabbit"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

func main() {
	// Setup event bus
	eventBus, err := setupEventBus()
	if err != nil {
		log.Fatalf("unable to setup event bus: %s", err)
	}

	// Run consumer
	go eventBus.Consume(context.Background(), nil, consumerFunc)

	// Run producer
	go runProducer(eventBus)

	// Run forever
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func setupEventBus() (*rabbit.Rabbit, error) {
	eventBus, err := rabbit.New(&rabbit.Options{
		URLs:      []string{"amqp://localhost"},
		QueueName: "my-queue",
		Bindings: []rabbit.Binding{
			{
				ExchangeName:    "events",
				ExchangeDeclare: true,
				ExchangeDurable: true,
				BindingKeys:     []string{"#"},
			},
		},
		QueueDurable:      true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new rabbit backend")
	}

	return eventBus, nil
}

func runProducer(eb *rabbit.Rabbit) {
	for {
		msg := fmt.Sprintf(`
		{
			"type": "new_order",
			"user_id": "%s"
		}`, uuid.NewV4().String())

		if err := eb.Publish(context.Background(), "foo", []byte(msg)); err != nil {
			log.Printf("unable to publish event: %s", err)
			continue
		}

		time.Sleep(5 * time.Second)
	}
}

func consumerFunc(msg amqp.Delivery) error {
	log.Printf("received new message on routing key '%s'", msg.RoutingKey)

	contents := make(map[string]interface{})

	if err := json.Unmarshal(msg.Body, &contents); err != nil {
		log.Printf("unable to decode msg: %s", err)
		return nil
	}

	msgType, ok := contents["type"]
	if !ok {
		log.Println("message does not contain 'type'")
		return nil
	}

	switch msgType {
	case "new_order":
		processNewOrder(contents)
	default:
		log.Printf("unknown msgType '%s' - skipping message", msgType)
		return nil
	}

	return nil
}

func processNewOrder(contents map[string]interface{}) {
	// Do stuff with the new order
	fmt.Printf("Doing something very clever with the contents: %+v\n", contents)
}
