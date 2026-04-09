package message

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/domain/entity"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

func StartDeleteOrderConsumer() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[StartDeleteOrderConsumer-1] Failed to connect to RabbitMQ: %v", err)
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[StartDeleteOrderConsumer-2] Failed to open a channel: %v", err)
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.NewConfig().PublisherName.ProductDelete,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[StartDeleteOrderConsumer-3] Failed to declare queue: %v", err)
	}

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
		log.Fatalf("[StartDeleteOrderConsumer-4] Failed to register consumer: %v", err)
	}

	log.Info("RabbitMQ Consumer started...")

	esClient, err := config.NewConfig().InitElasticsearch()
	if err != nil {
		log.Errorf("[StartDeleteOrderConsumer-5] Failed initialize Elasticsearch client: %v", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			var data map[string]string
			err := json.Unmarshal(d.Body, &data)
			if err != nil {
				log.Errorf("[StartDeleteOrderConsumer-6] Error decoding message: %v", err)
				continue
			}

			productID := data["ProductID"]

			res, err := esClient.Delete("products", productID)
			if err != nil {
				log.Errorf("[StartDeleteOrderConsumer-8] Error indexing to Elasticsearch: %v", err)
				continue
			}
			defer res.Body.Close()
		}
	}()

	log.Infof("[StartDeleteOrderConsumer-10] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func StartConsumer() {

	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[StartConsumer-1] Gagal koneksi ke RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[StartConsumer-2] Gagal membuka channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.NewConfig().PublisherName.ProductPublish,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[StartConsumer-3] Gagal deklarasi queue: %v", err)
	}

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
		log.Fatalf("[StartConsumer-4] Gagal register consumer: %v", err)
	}

	esClient, err := config.NewConfig().InitElasticsearch()
	if err != nil {
		log.Errorf("[StartConsumer-5] Gagal inisialisasi Elasticsearch: %v", err)
		return
	}

	log.Infof("[StartConsumer-6] Worker RabbitMQ berjalan, menunggu pesan...")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			var product entity.ProductEntity
			if err := json.Unmarshal(d.Body, &product); err != nil {
				log.Errorf("[StartConsumer-7] Gagal unmarshal pesan: %v", err)
				d.Nack(false, true)
				continue
			}

			res, err := esClient.Index(
				"products",
				bytes.NewReader(d.Body),
				esClient.Index.WithDocumentID(uuid.UUID(product.ID).String()),
				esClient.Index.WithContext(context.Background()),
				esClient.Index.WithRefresh("true"),
			)
			if err != nil {
				log.Errorf("[StartConsumer-8] Gagal indexing ke Elasticsearch: %v", err)
				d.Nack(false, true)
				continue
			}
			defer res.Body.Close()

			if err := d.Ack(false); err != nil {
				log.Errorf("[StartConsumer-9] Gagal ack pesan: %v", err)
				continue
			}

			log.Infof("[StartConsumer-10] Produk %s berhasil diindex ke Elasticsearch", product.ID)
		}
	}()

	<-forever
}
