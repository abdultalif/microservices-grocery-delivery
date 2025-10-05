package messaging

import (
	"fmt"

	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Cfg config.RabbitMQ
}

func NewRabbitMQ(cfg config.RabbitMQ) *RabbitMQ {
	return &RabbitMQ{Cfg: cfg}
}

func (r *RabbitMQ) Connect() (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		r.Cfg.User,
		r.Cfg.Password,
		r.Cfg.Host,
		r.Cfg.Port,
		r.Cfg.VirtualHost)

	debugUrl := fmt.Sprintf("amqp://%s:***@%s:%s/%s",
		r.Cfg.User,
		r.Cfg.Host,
		r.Cfg.Port,
		r.Cfg.VirtualHost)
	log.Infof("Attempting to connect to RabbitMQ: %s", debugUrl)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Errorf("Failed to connect to RabbitMQ with URL: %s, Error: %v", debugUrl, err)
		return nil, err
	}

	log.Info("Successfully connected to RabbitMQ")
	return conn, nil
}
