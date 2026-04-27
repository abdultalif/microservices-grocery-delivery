package config

import "github.com/spf13/viper"

type App struct {
	AppPort           string `json:"app_port"`
	AppEnv            string `json:"app_env"`
	JwtSecret         string `json:"jwt_secret"`
	ServerTimeOut     int    `json:"server_timeout"`
	ProductServiceUrl string `json:"product_service_url"`
	APIGatewayUrl     string `json:"api_gateway_url"`
	UserServiceUrl    string `json:"user_service_url"`
	OrderServiceUrl   string `json:"order_service_url"`
	AuthClientID      string `json:"auth_client_id"`
	AuthClientSecret  string `json:"auth_client_secret"`
}

type PostgresDB struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"db_name"`
	DBMaxOpen int    `json:"db_max_open"`
	DBMaxIdle int    `json:"db_max_idle"`
}

type Redis struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RabbitMQ struct {
	Host        string `json:"host"`
	User        string `json:"user"`
	Password    string `json:"password"`
	VirtualHost string `json:"virtual_host"`
	Port        string `json:"port"`
}

type Midtrans struct {
	ServerKey string `json:"server_key"`
}

type PublisherName struct {
	PublisherPaymentSuccess string `json:"payment_success"`
}

type Config struct {
	App           App           `json:"app"`
	Postgres      PostgresDB    `json:"postgres"`
	Redis         Redis         `json:"redis"`
	RabbitMQ      RabbitMQ      `json:"rabbitmq"`
	Midtrans      Midtrans      `json:"midtrans"`
	PublisherName PublisherName `json:"publisher_name"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort:           viper.GetString("APP_PORT"),
			AppEnv:            viper.GetString("APP_ENV"),
			JwtSecret:         viper.GetString("JWT_SECRET"),
			ServerTimeOut:     viper.GetInt("SERVER_TIMEOUT"),
			ProductServiceUrl: viper.GetString("PRODUCT_SERVICE_URL"),
			UserServiceUrl:    viper.GetString("USER_SERVICE_URL"),
			OrderServiceUrl:   viper.GetString("ORDER_SERVICE_URL"),
			APIGatewayUrl:     viper.GetString("API_GATEWAY_URL"),
			AuthClientID:      viper.GetString("AUTH_CLIENT_ID"),
			AuthClientSecret:  viper.GetString("AUTH_CLIENT_SECRET"),
		},
		Postgres: PostgresDB{
			Host:      viper.GetString("DATABASE_HOST"),
			Port:      viper.GetString("DATABASE_PORT"),
			User:      viper.GetString("DATABASE_USER"),
			Password:  viper.GetString("DATABASE_PASSWORD"),
			DBName:    viper.GetString("DATABASE_NAME"),
			DBMaxOpen: viper.GetInt("DATABASE_MAX_OPEN_CONNECTION"),
			DBMaxIdle: viper.GetInt("DATABASE_MAX_IDLE_CONNECTION"),
		},
		Redis: Redis{
			Host: viper.GetString("REDIS_HOST"),
			Port: viper.GetString("REDIS_PORT"),
		},
		RabbitMQ: RabbitMQ{
			Host:        viper.GetString("RABBITMQ_HOST"),
			Port:        viper.GetString("RABBITMQ_PORT"),
			User:        viper.GetString("RABBITMQ_USER"),
			Password:    viper.GetString("RABBITMQ_PASSWORD"),
			VirtualHost: viper.GetString("RABBITMQ_VIRTUAL_HOST"),
		},
		Midtrans: Midtrans{
			ServerKey: viper.GetString("MIDTRANS_SERVER_KEY"),
		},
		PublisherName: PublisherName{
			PublisherPaymentSuccess: viper.GetString("PUBLISHER_PAYMENT_SUCCESS"),
		},
	}
}
