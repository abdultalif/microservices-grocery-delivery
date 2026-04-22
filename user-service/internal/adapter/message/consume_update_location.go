package message

import (
	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/domain/model"
	"github.com/labstack/gommon/log"
)

func ConsumeUpdateUserLocation() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[ConsumeUpdateUserLocation-1] Failed to connect RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeUpdateUserLocation-2] Failed to open channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.NewConfig().PublisherName.UserLocationUpdate,
		true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("[ConsumeUpdateUserLocation-3] Failed to declare queue: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[ConsumeUpdateUserLocation-4] Failed to register consumer: %v", err)
		return
	}

	db, err := config.NewConfig().ConnectionPostgres()
	if err != nil {
		log.Errorf("[ConsumeUpdateUserLocation-5] Failed to connect DB: %v", err)
		return
	}

	log.Info("RabbitMQ Consumer user.location.update started...")

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			var payload struct {
				UserID int64  `json:"user_id"`
				Lat    string `json:"lat"`
				Lng    string `json:"lng"`
			}

			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Errorf("[ConsumeUpdateUserLocation-6] Failed to parse message: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Infof("[ConsumeUpdateUserLocation] Received payload: userID=%d lat=%s lng=%s",
				payload.UserID, payload.Lat, payload.Lng)

			// Update lat & lng di tabel users milik user service
			if err := db.DB.Model(&model.User{}).
				Where("id = ?", payload.UserID).
				Updates(map[string]interface{}{
					"lat": payload.Lat,
					"lng": payload.Lng,
				}).Error; err != nil {
				log.Errorf("[ConsumeUpdateUserLocation-7] Failed to update location user %d: %v",
					payload.UserID, err)
				msg.Nack(false, true)
				continue
			}

			log.Infof("[ConsumeUpdateUserLocation-8] Location updated for user %d", payload.UserID)
			msg.Ack(false)
		}
	}()

	log.Info("[ConsumeUpdateUserLocation] Waiting for messages. To exit press CTRL+C")
	<-forever
}
