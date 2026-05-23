package config

import (
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

const (
	maxRetries    = 5
	retryInterval = 3 * time.Second
)

func (cfg Config) NewRabbitMQ() (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.VirtualHost,
	)

	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			log.Printf("[NewRabbitMQ-1]Successfully connected to RabbitMQ on attempt %d\n", i)
			return conn, nil
		}

		log.Warnf("[NewRabbitMQ] Attempt %d/%d failed: %v. Retrying in %s...",
			i, maxRetries, err, retryInterval)

		if i < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("[NewRabbitMQ] failed after %d attempts: %w", maxRetries, err)

}
