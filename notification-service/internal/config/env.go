package config

import (
	"log"

	"github.com/spf13/viper"
)

type Env struct {
	AppPort   string
	AppEnv    string
	JwtSecret string

	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseMaxOpen  int
	DatabaseMaxIdle  int

	RabbitHost        string
	RabbitPort        string
	RabbitUser        string
	RabbitPassword    string
	RabbitVirtualHost string

	EmailHost string
	EmailPort int
	EmailUser string
	EmailPass string
	EmailTls  bool
	EmailFrom string
}

func LoadEnv() *Env {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	return &Env{
		AppPort:   viper.GetString("APP_PORT"),
		AppEnv:    viper.GetString("APP_ENV"),
		JwtSecret: viper.GetString("JWT_SECRET"),

		DatabaseHost:     viper.GetString("DATABASE_HOST"),
		DatabasePort:     viper.GetString("DATABASE_PORT"),
		DatabaseUser:     viper.GetString("DATABASE_USER"),
		DatabasePassword: viper.GetString("DATABASE_PASSWORD"),
		DatabaseName:     viper.GetString("DATABASE_NAME"),
		DatabaseMaxOpen:  viper.GetInt("DATABASE_MAX_OPEN_CONNECTION"),
		DatabaseMaxIdle:  viper.GetInt("DATABASE_MAX_IDLE_CONNECTION"),

		RabbitHost:        viper.GetString("RABBITMQ_HOST"),
		RabbitPort:        viper.GetString("RABBITMQ_PORT"),
		RabbitUser:        viper.GetString("RABBITMQ_USER"),
		RabbitPassword:    viper.GetString("RABBITMQ_PASSWORD"),
		RabbitVirtualHost: viper.GetString("RABBITMQ_VIRTUAL_HOST"),

		EmailHost: viper.GetString("EMAIL_HOST"),
		EmailPort: viper.GetInt("EMAIL_PORT"),
		EmailUser: viper.GetString("EMAIL_USER"),
		EmailPass: viper.GetString("EMAIL_PASSWORD"),
		EmailTls:  viper.GetBool("EMAIL_TLS"),
		EmailFrom: viper.GetString("EMAIL_FROM"),
	}
}
