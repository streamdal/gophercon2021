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
	ServiceName string
	Deps        *Dependencies
}

// Event describes what a valid event looks like. Normally, this should be shared
// between services, best in the form of compiled protos that are vendored.
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

		if err := c.Deps.StatsDClient.Incr(c.ServiceName+"_new_order_skipped", nil, 1); err != nil {
			log.Errorf("unable to increase count: %s", err)
		}

		return nil
	}

	if err := c.Deps.StatsDClient.Incr(c.ServiceName+"_new_order_ok", nil, 1); err != nil {
		log.Errorf("unable to increase count: %s", err)
	}

	log.Infof("successfully processed 'new_order' message with id '%s'", msg.ID)

	// Update state
	c.Deps.State.Add(msg.ID)

	return nil
}

func (c *Consumer) banUser(msg *Event) error {
	log.Info("Processing ban_user message")

	if err := c.Deps.StatsDClient.Incr(c.ServiceName+"_ban_user_ok", nil, 1); err != nil {
		log.Errorf("unable to increase count: %s", err)
	}

	return nil
}
