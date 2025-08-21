package message

// import (
// 	"order-service/config"

// 	"github.com/google/uuid"
// 	"github.com/labstack/gommon/log"
// )

// type PublishRabbitMQInterface interface {
// 	PublishUpdateStock(productID uuid.UUID, quantity int64)
// }

// type PublishRabbitMQ struct {
// 	cfg *config.Config
// }

// // PublishUpdateStock implements PublishRabbitMQInterface.
// func (p *PublishRabbitMQ) PublishUpdateStock(productID uuid.UUID, quantity int64) {

// 	conn, err := p.cfg.NewRabbitMQ()
// 	if err != nil {
// 		log.Errorf("[PublishUpdateStock-1] Failed to connect to RabbitMQ: %v", err)
// 		return
// 	}

// 	defer conn.Close()

// 	ch, err := conn.Channel()
// 	if err != nil {
// 		log.Errorf("[PublishUpdateStock-2] Failed to open a channel: %v", err)
// 		return
// 	}

// 	defer ch.Close()

// 	q, err := ch.

// }

// func NewPublisherRabbitMQ(cfg *config.Config) PublishRabbitMQInterface {
// 	return &PublishRabbitMQ{
// 		cfg: cfg,
// 	}
// }
