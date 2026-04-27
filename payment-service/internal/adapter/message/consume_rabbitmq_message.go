package message

import (
	"github.com/abdultalif/microservices-grocery-delivery/payment-service/config"
	"github.com/labstack/gommon/log"
)

func ConsumeUserEventForPayment() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeUserEventForPayment-1] RMQ error: %v", err)
		return 
	}
	defer conn.Close()
	
	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeUserEventForPayment-1] Channel error: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"user.events"
	)
}