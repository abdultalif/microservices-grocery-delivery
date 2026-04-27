package message

import (
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type UserEvent struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
}

func PublishUserEvent(eventType string, payload interface{}) error {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishUserEvent-1] Failed connect RMQ: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishUserEvent-2] Failed open channel: %v", err)
		return err
	}
	defer ch.Close()

	exchangeName := "user.events"

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishUserEvent-3] Failed declare exchange: %v", err)
		return err
	}

	event := UserEvent{
		EventType: eventType,
		Data:      payload,
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Errorf("[PublishUserEvent-4] Marshal error: %v", err)
		return err
	}

	err = ch.Publish(
		exchangeName,
		eventType,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Errorf("[PublishUserEvent-5] Publish error: %v", err)
		return err
	}

	log.Infof("[UserEvent] Published %s", eventType)
	return nil
}
