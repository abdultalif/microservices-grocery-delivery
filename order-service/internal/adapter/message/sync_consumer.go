package message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abdultalif/microservices-grocery-delivery/order-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/internal/core/domain/entity"
	"github.com/abdultalif/microservices-grocery-delivery/order-service/utils"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

const (
	exchangeName   = "user.events"
	dlxName        = "user.events.dlx" // Dead Letter Exchange
	queueName      = "user.events.order"
	dlqName        = "user.events.order.dlq" // Dead Letter Queue
	prefetchCount  = 10
	reconnectDelay = 5 * time.Second
)

type Event struct {
	EventType string                 `json:"event_type"`
	Data      entity.ProductSnapshot `json:"data"`
}

// userEventPayload adalah struktur pesan yang diterima dari user-service.
type userEventPayload struct {
	EventType string                        `json:"event_type"`
	Data      entity.CustomerResponseEntity `json:"data"`
}

// ConsumeUserEvents memulai consumer loop dengan reconnect otomatis.
// Berhenti saat ctx dibatalkan (dipanggil dari app.go saat shutdown).
func ConsumeUserEvents(ctx context.Context, cfg config.Config) {
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[ConsumeUserEvents] Failed to connect postgres: %v", err)
	}
	localRepo := repository.NewLocalDataRepository(db.DB)

	// Loop ini memastikan consumer selalu berjalan meski RabbitMQ sempat mati.
	for {
		select {
		case <-ctx.Done():
			log.Info("[ConsumeUserEvents] Context cancelled, stopping consumer.")
			return
		default:
		}

		log.Info("[ConsumeUserEvents] Connecting to RabbitMQ...")
		if err := runConsumer(ctx, cfg, localRepo); err != nil {
			log.Errorf("[ConsumeUserEvents] Consumer error: %v. Reconnecting in %s...", err, reconnectDelay)
		}

		// Tunggu sebelum reconnect, tapi tetap responsive terhadap ctx.Done().
		select {
		case <-ctx.Done():
			log.Info("[ConsumeUserEvents] Shutdown signal received during backoff.")
			return
		case <-time.After(reconnectDelay):
		}
	}
}

// runConsumer menjalankan satu sesi consumer hingga koneksi terputus atau ctx dibatalkan.
func runConsumer(ctx context.Context, cfg config.Config, localRepo repository.LocalDataRepositoryInterface) error {
	conn, err := cfg.NewRabbitMQ()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// Pantau jika koneksi tiba-tiba terputus dari sisi broker.
	connClosed := make(chan *amqp.Error, 1)
	conn.NotifyClose(connClosed)

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	// Setup Dead Letter Exchange (DLX): pesan yang di-Nack tanpa requeue
	// akan dikirim ke sini, sehingga bisa diinspeksi tanpa mengganggu queue utama.
	if err := ch.ExchangeDeclare(dlxName, "fanout", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare DLX: %w", err)
	}
	if _, err := ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare DLQ: %w", err)
	}
	if err := ch.QueueBind(dlqName, "#", dlxName, false, nil); err != nil {
		return fmt.Errorf("bind DLQ: %w", err)
	}

	// Exchange utama
	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	// Queue utama dengan x-dead-letter-exchange: pesan gagal otomatis masuk DLX.
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": dlxName, // pesan Nack tanpa requeue → DLQ
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(q.Name, "user.*", exchangeName, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	// Prefetch: consumer hanya ambil N pesan sekaligus, tidak membanjiri memory.
	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("set QoS: %w", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}

	log.Infof("[ConsumeUserEvents] Consumer '%s' started", queueName)

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown: hentikan loop, pesan yang sedang diproses selesai dulu.
			log.Info("[ConsumeUserEvents] Shutting down consumer gracefully...")
			return nil

		case amqpErr := <-connClosed:
			return fmt.Errorf("connection closed by broker: %v", amqpErr)

		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}
			handleMessage(ctx, msg, localRepo)
		}
	}
}

// handleMessage memproses satu pesan. Nack tanpa requeue jika payload corrupt
// (tidak bisa di-retry), Nack dengan requeue jika error sementara (DB down, dll).
func handleMessage(ctx context.Context, msg amqp.Delivery, localRepo repository.LocalDataRepositoryInterface) {
	var event userEventPayload
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		// Payload rusak → tidak ada gunanya di-requeue, kirim ke DLQ.
		log.Errorf("[handleMessage] Unmarshal error (sending to DLQ): %v", err)
		msg.Nack(false, false) // requeue=false → masuk DLQ via x-dead-letter-exchange
		return
	}

	var processErr error
	switch event.EventType {
	case "user.created", "user.updated":
		processErr = localRepo.UpsertBuyer(ctx, event.Data)
	case "user.deleted":
		processErr = localRepo.DeleteBuyer(ctx, event.Data.ID)
	default:
		// Event tidak dikenal → buang saja, jangan requeue selamanya.
		log.Warnf("[handleMessage] Unknown event type '%s', discarding", event.EventType)
		msg.Nack(false, false)
		return
	}

	if processErr != nil {
		// Error sementara (misalnya DB timeout) → requeue agar dicoba lagi.
		log.Errorf("[handleMessage] Process error for event '%s': %v. Requeuing...", event.EventType, processErr)
		msg.Nack(false, true) // requeue=true
		return
	}

	msg.Ack(false)
	log.Infof("[handleMessage] Event '%s' processed successfully (ID: %s)", event.EventType, event.Data.ID)
}

func ConsumeProductEvents() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeProductEvents-1] Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeProductEvents-2] Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		utils.PRODUCT_EXCHANGE,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-3] Failed declare exchange: %v", err)
		return
	}

	q, err := ch.QueueDeclare(
		utils.ORDER_PRODUCT_QUEUE,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-4] Failed to declare queue: %v", err)
		return
	}

	err = ch.QueueBind(
		q.Name,
		"product.*",
		utils.PRODUCT_EXCHANGE,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-5] Failed bind queue: %v", err)
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-6] Failed set QoS: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Errorf("[ConsumeProductEvents-7] Failed to register consumer: %v", err)
		return
	}

	db, _ := config.NewConfig().ConnectionPostgres()
	localDataRepo := repository.NewLocalDataRepository(db.DB)

	log.Infof("[ConsumeProductEvents-8] Listening on queue: %s", q.Name)

	for msg := range msgs {

		var event Event

		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Errorf("[ConsumeProductEvents-9] Failed to unmarshal message: %v", err)
			msg.Ack(false)
			continue
		}

		var processErr error

		switch msg.RoutingKey {

		case utils.PRODUCT_CREATED_RK, utils.PRODUCT_UPDATED_RK:

			parent := entity.ProductSnapshot{
				ID:           event.Data.ID,
				Name:         event.Data.Name,
				Stock:        event.Data.Stock,
				Image:        event.Data.Image,
				RegulerPrice: event.Data.RegulerPrice,
				SalePrice:    event.Data.SalePrice,
				Unit:         event.Data.Unit,
				Weight:       event.Data.Weight,
				CreatedAt:    event.Data.CreatedAt,
			}

			processErr = localDataRepo.UpsertProduct(context.Background(), parent)
			if processErr != nil {
				break
			}

			for _, child := range event.Data.Child {

				childData := entity.ProductSnapshot{
					ID:           child.ID,
					ParentID:     &event.Data.ID,
					Name:         event.Data.Name,
					Stock:        child.Stock,
					Image:        child.Image,
					RegulerPrice: child.RegulerPrice,
					SalePrice:    child.SalePrice,
					Unit:         child.Unit,
					Weight:       child.Weight,
					CreatedAt:    child.CreatedAt,
				}

				processErr = localDataRepo.UpsertProduct(context.Background(), childData)
				if processErr != nil {
					break
				}
			}

		case utils.PRODUCT_DELETED_RK:

			processErr = localDataRepo.DeleteProduct(context.Background(), event.Data.ID)
		}

		if processErr != nil {
			log.Errorf("[ConsumeProductEvents-10] Process error: %v", processErr)

			msg.Ack(false)
			continue
		}

		log.Infof("[ConsumeProductEvents-11] Product %s processed (%s)", event.Data.ID, msg.RoutingKey)
		msg.Ack(false)
	}
}
