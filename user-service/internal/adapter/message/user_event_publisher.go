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

	queueName := "user.events"

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishUserEvent-3] Failed declare queue: %v", err)
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
		"",
		queueName,
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
