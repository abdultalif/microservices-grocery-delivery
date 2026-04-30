package messaging

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/config"

	"github.com/go-mail/mail"
	"github.com/labstack/gommon/log"
)

type MessageEmailInterface interface {
	SendEmailNotification(to, subject, body string) error
}

type SendAttribute struct {
	Username string
	Password string
	Host     string
	Port     int
	From     string
	IsTLS    bool
}

// SendEmailNotification implements MessageEmailInterface.
func (s *SendAttribute) SendEmailNotification(to string, subject string, body string) error {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)

	log.Infof("SendEmailNotification called with - To: '%s', Subject: '%s', Body length: %d", to, subject, len(body))

	if to == "" {
		err := fmt.Errorf("recipient email address is empty")
		log.Errorf("[SendEmailNotification] %v", err)
		return err
	}

	if subject == "" {
		subject = "No Subject"
		log.Warnf("Subject is empty, using default: %s", subject)
	}

	if body == "" {
		body = "No content"
		log.Warnf("Body is empty, using default: %s", body)
	}

	if s.From == "" {
		err := fmt.Errorf("sender email address (From) is not configured")
		log.Errorf("[SendEmailNotification] %v", err)
		return err
	}

	if s.Username == "" || s.Password == "" {
		err := fmt.Errorf("email credentials are not configured - Username: '%s', Password length: %d", s.Username, len(s.Password))
		log.Errorf("[SendEmailNotification] %v", err)
		return err
	}

	log.Infof("Email config - Host: %s, Port: %d, From: %s, Username: %s, TLS: %v",
		s.Host, s.Port, s.From, s.Username, s.IsTLS)

	m := mail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(s.Host, s.Port, s.Username, s.Password)

	d.TLSConfig = &tls.Config{
		ServerName: s.Host,
	}

	if s.IsTLS {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	} else {
		d.StartTLSPolicy = mail.NoStartTLS
	}

	log.Infof("Dialing SMTP: %s:%d with user %s", s.Host, s.Port, s.Username)

	if err := d.DialAndSend(m); err != nil {
		log.Errorf("[SendEmailNotification] Failed to send email to %s: %v", to, err)
		return err
	}

	log.Infof("Email sent successfully to: %s", to)
	return nil
}

func NewMessageEmail(cfg *config.Config) MessageEmailInterface {
	log.Infof("Creating email service with config - Host: %s, Port: %d, User: %s, From: %s, TLS: %v",
		cfg.Email.Host, cfg.Email.Port, cfg.Email.User, cfg.Email.From, cfg.Email.Tls)

	return &SendAttribute{
		Username: cfg.Email.User,
		Password: cfg.Email.Pass,
		Host:     cfg.Email.Host,
		Port:     cfg.Email.Port,
		From:     cfg.Email.From,
		IsTLS:    cfg.Email.Tls,
	}
}
