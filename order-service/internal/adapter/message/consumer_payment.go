package message

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

const (
	ExchangePaymentEvents = "payment.events"
)

func ConsumePaymentEvent() {

	cfg := config.NewConfig()

	// rabbitmq connection
	conn, err := cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-1] rabbitmq: %v", err)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-2] channel: %v", err)
		return
	}

	defer ch.Close()

	// exchange
	err = ch.ExchangeDeclare(
		ExchangePaymentEvents,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-3] exchange: %v", err)
		return
	}

	// queue
	q, err := ch.QueueDeclare(
		"order.payment.events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-4] queue: %v", err)
		return
	}

	// binding routing key
	routingKeys := []string{
		"payment.paid",
		"payment.cancelled",
		"payment.refunded",
	}

	for _, routingKey := range routingKeys {

		err = ch.QueueBind(
			q.Name,
			routingKey,
			ExchangePaymentEvents,
			false,
			nil,
		)
		if err != nil {

			log.Errorf(
				"[ConsumePaymentEvent-5] bind: %v",
				err,
			)

			return
		}
	}

	// consumer
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-6] consume: %v", err)
		return
	}

	// elasticsearch
	esClient, err := cfg.InitElasticsearch()
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-7] elasticsearch: %v", err)
		return
	}

	// postgres
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Errorf("[ConsumePaymentEvent-8] postgres: %v", err)
		return
	}

	orderRepo := repository.NewOrderRepository(
		db.DB,
	)

	log.Info("[ConsumePaymentEvent] worker started...")

	forever := make(chan bool)

	go func() {

		for msg := range msgs {

			var payload entity.PaymentEvent

			err := json.Unmarshal(
				msg.Body,
				&payload,
			)
			if err != nil {

				log.Errorf(
					"[ConsumePaymentEvent-9] unmarshal: %v",
					err,
				)

				msg.Nack(false, false)

				continue
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			orderStatus := mapPaymentStatus(
				payload.Status,
			)

			remarks := "payment " + payload.Status

			// update postgres
			_, _, _, err = orderRepo.UpdateStatus(
				ctx,
				entity.OrderEntity{
					OrderCode: payload.OrderCode,
					Status:    orderStatus,
					Remarks:   remarks,
				},
			)

			cancel()

			if err != nil {

				log.Errorf(
					"[ConsumePaymentEvent-10] postgres: %v",
					err,
				)

				msg.Nack(false, true)

				continue
			}

			// update elasticsearch
			updateScript := map[string]interface{}{
				"script": map[string]interface{}{
					"source": `
						ctx._source.PaymentMethod = params.payment_method;
						ctx._source.Status = params.status;
					`,
					"lang": "painless",
					"params": map[string]interface{}{
						"payment_method": payload.PaymentMethod,
						"status":         orderStatus,
					},
				},
			}

			body, err := json.Marshal(updateScript)
			if err != nil {

				log.Errorf(
					"[ConsumePaymentEvent-11] marshal elastic: %v",
					err,
				)

				msg.Nack(false, true)

				continue
			}

			res, err := esClient.Update(
				"orders",
				payload.OrderCode,
				bytes.NewReader(body),
			)
			if err != nil {

				log.Errorf(
					"[ConsumePaymentEvent-12] elastic update: %v",
					err,
				)

				msg.Nack(false, true)

				continue
			}

			defer res.Body.Close()

			responseBody, _ := io.ReadAll(
				res.Body,
			)

			log.Infof(
				"[ConsumePaymentEvent] elastic response: %s",
				string(responseBody),
			)

			err = msg.Ack(false)
			if err != nil {

				log.Errorf(
					"[ConsumePaymentEvent-13] ack: %v",
					err,
				)

				continue
			}

			log.Infof(
				"[ConsumePaymentEvent] success orderCode=%s status=%s",
				payload.OrderCode,
				payload.Status,
			)
		}
	}()

	log.Info("[ConsumePaymentEvent] waiting message...")
	<-forever
}

func mapPaymentStatus(
	status string,
) string {

	switch status {

	case "paid":
		return "Confirmed"

	case "cancelled":
		return "Cancelled"

	case "refunded":
		return "Refunded"

	default:
		return "Pending"
	}
}
