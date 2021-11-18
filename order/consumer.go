package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	NewOrder MessageType = "new_order"
	BanUser  MessageType = "ban_user"
)

type MessageType string

type Consumer struct {
	Deps *Dependencies
}

type Event struct {
	ID   string
	Type MessageType
	Args map[string]interface{}
}

func (c *Consumer) ConsumeFunc(msg amqp.Delivery) error {
	log.Printf("received msg on routing key '%s'\n", msg.RoutingKey)

	// Try to decode the msg
	decoded := &Event{}

	if err := json.Unmarshal(msg.Body, decoded); err != nil {
		log.Errorf("unable to decode message: %s", err)
		return nil
	}

	var err error

	switch decoded.Type {
	case NewOrder:
		err = c.processOrder(decoded)
	case BanUser:
		err = c.banUser(decoded)
	default:
		log.Warningf("unknown message type '%s'", decoded.Type)
		return nil
	}

	if err != nil {
		log.Errorf("unable to handle message '%s' with id '%s': %s", decoded.Type, decoded.ID, err)
		return nil
	}

	return nil
}

func (c *Consumer) processOrder(msg *Event) error {
	log.Info("Processing new_order message")

	// Ensure this is not something we've already processed
	if c.Deps.State.Contains(msg.ID) {
		log.Infof("message id '%s' is a dupe - skipping", msg.ID)

		c.Deps.StatsDClient.Count("order_processed_dupe", 1, nil, 1)
		return nil
	}

	c.Deps.StatsDClient.Count("order_processed", 1, nil, 1)

	log.Infof("successfully processed message with id '%s'", msg.ID)

	// Update state
	c.Deps.State.Add(msg.ID)

	return nil
}

func (c *Consumer) banUser(msg *Event) error {
	c.Deps.StatsDClient.Count("user_banned", 1, nil, 1)

	return nil
}
