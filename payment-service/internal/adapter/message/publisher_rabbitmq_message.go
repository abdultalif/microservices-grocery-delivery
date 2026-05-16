package message

import (
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/payment-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

const (
	ExchangePaymentEvents = "payment.events"

	RoutingKeyPaymentPaid      = "payment.paid"
	RoutingKeyPaymentCancelled = "payment.cancelled"
	RoutingKeyPaymentRefunded  = "payment.refunded"
)

type PublishRabbitMQInterface interface {
	PublishPaymentEvent(routingKey string, payment entity.PaymentEvent) error
}

type PublishRabbitMQ struct {
	channel *amqp.Channel
}

func (p *PublishRabbitMQ) setupExchange() error {

	return p.channel.ExchangeDeclare(
		ExchangePaymentEvents,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

// PublishPaymentEvent implements PublishRabbitMQInterface.
func (p *PublishRabbitMQ) PublishPaymentEvent(routingKey string, payment entity.PaymentEvent) error {
	err := p.setupExchange()
	if err != nil {
		log.Errorf("[PublishPaymentEvent-1] failed setup exchange: %v", err)
		return err
	}

	body, err := json.Marshal(payment)
	if err != nil {
		log.Errorf("[PublishPaymentEvent-2] failed marshal payload: %v", err)
		return err
	}

	err = p.channel.Publish(
		ExchangePaymentEvents,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		log.Errorf("[PublishPaymentEvent-3] failed publish message: %v", err)
		return err
	}

	log.Infof(
		"[PublishPaymentEvent] success publish event: %s",
		routingKey,
	)

	return nil
}

func NewPublishRabbitMQ(channel *amqp.Channel) PublishRabbitMQInterface {

	return &PublishRabbitMQ{
		channel: channel,
	}

}
