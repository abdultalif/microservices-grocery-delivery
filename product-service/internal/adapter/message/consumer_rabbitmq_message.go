package message

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"product-service/config"
	esadapter "product-service/internal/adapter/elasticsearch"
	"product-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

func StartConsumer() {
	cfg := config.NewConfig()

	conn, err := cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("RabbitMQ connection failed: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Channel creation failed: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("elasticsearch_indexing", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Queue declare failed: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Queue consume failed: %v", err)
		return
	}

	esClient, err := config.InitElasticsearch()
	if err != nil {
		log.Errorf("Elasticsearch client init failed: %v", err)
		return
	}

	if err := esadapter.EnsureProductIndexExists(esClient); err != nil {
		log.Errorf("Ensure index exists failed: %v", err)
		return
	}

	go func() {
		for d := range msgs {
			var product entity.ProductEntity
			if err := json.Unmarshal(d.Body, &product); err != nil {
				log.Errorf("Unmarshal failed: %v", err)
				continue
			}

			body, err := json.Marshal(product)
			if err != nil {
				log.Errorf("Marshal failed: %v", err)
				continue
			}

			res, err := esClient.Index(
				"products",
				bytes.NewReader(body),
				esClient.Index.WithDocumentID(product.ID.String()),
				esClient.Index.WithContext(context.Background()),
				esClient.Index.WithRefresh("true"),
			)
			if err != nil {
				log.Errorf("Indexing failed: %v", err)
				continue
			}
			defer res.Body.Close()

			respBody, _ := io.ReadAll(res.Body)
			log.Infof("Indexed product: %s", respBody)
		}
	}()

	log.Info("Waiting for messages...")
	select {}
}
