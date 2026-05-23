package message

import (
	"fmt"
	"sync"
	"time"

	"encoding/json"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

const (
	exchangeName   = "user.events"
	publishTimeout = 5 * time.Second
)

// UserEvent adalah struktur event yang dikirim ke RabbitMQ.
type UserEvent struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
}

// UserEventPublisher mengelola koneksi dan channel RabbitMQ secara persisten.
// Gunakan NewUserEventPublisher() untuk membuat instance, lalu inject ke service.
type UserEventPublisher struct {
	mu   sync.Mutex
	cfg  config.Config
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewUserEventPublisher(cfg config.Config) (*UserEventPublisher, error) {
	p := &UserEventPublisher{cfg: cfg}
	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

// connect membangun koneksi dan channel, lalu mendeklarasikan exchange.
// Dipanggil saat init dan saat reconnect otomatis.
func (p *UserEventPublisher) connect() error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		return fmt.Errorf("[Publisher] connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("[Publisher] open channel: %w", err)
	}

	// Publisher confirm: broker konfirmasi setiap pesan yang diterima.
	// Ini memastikan pesan tidak hilang tanpa kita tahu.
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("[Publisher] confirm mode: %w", err)
	}

	if err := ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable: exchange bertahan saat RabbitMQ restart
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("[Publisher] declare exchange: %w", err)
	}

	p.conn = conn
	p.ch = ch
	log.Infof("[Publisher] RabbitMQ connection established")
	return nil
}

// reconnect mencoba membangun ulang koneksi jika terputus.
func (p *UserEventPublisher) reconnect() error {
	log.Warn("[Publisher] Attempting reconnect to RabbitMQ...")
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	return p.connect()
}

// Publish mengirim event ke RabbitMQ dengan routing key = eventType.
// Otomatis reconnect jika koneksi terputus.
// Thread-safe.
func (p *UserEventPublisher) Publish(eventType string, payload interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(UserEvent{
		EventType: eventType,
		Data:      payload,
	})
	if err != nil {
		return fmt.Errorf("[Publisher] marshal error: %w", err)
	}

	// Coba publish; jika gagal karena koneksi mati, reconnect sekali lalu retry.
	err = p.publish(eventType, body)
	if err != nil {
		log.Warnf("[Publisher] Publish failed (%v), reconnecting...", err)
		if reconnErr := p.reconnect(); reconnErr != nil {
			return fmt.Errorf("[Publisher] reconnect failed: %w", reconnErr)
		}
		err = p.publish(eventType, body)
	}

	return err
}

// publish adalah operasi publish internal tanpa lock.
func (p *UserEventPublisher) publish(routingKey string, body []byte) error {
	confirms := p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := p.ch.Publish(
		exchangeName,
		routingKey,
		true, // mandatory: kembalikan error jika tidak ada queue yang cocok
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // pesan bertahan saat RabbitMQ restart
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("ch.Publish: %w", err)
	}

	// Tunggu konfirmasi dari broker (publisher confirm mode).
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("broker nack'd the message (tag=%d)", confirm.DeliveryTag)
		}
		log.Infof("[Publisher] Event '%s' confirmed by broker", routingKey)
		return nil
	case <-time.After(publishTimeout):
		return fmt.Errorf("timeout waiting for broker confirm on '%s'", routingKey)
	}
}

// Close menutup koneksi dengan bersih saat aplikasi shutdown.
func (p *UserEventPublisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Info("[Publisher] RabbitMQ connection closed")
}
