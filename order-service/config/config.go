package config

import "github.com/spf13/viper"

type App struct {
	AppPort       string `json:"app_port"`
	AppEnv        string `json:"app_env"`
	JwtSecret     string `json:"jwt_secret"`
	ServerTimeOut int    `json:"server_timeout"`

	ApiGatewayServiceUrl string `json:"api_gateway_service_url"`
	ProductServiceUrl    string `json:"product_service_url"`
	UserServiceUrl       string `json:"user_service_url"`

	AuthClientID     string `json:"auth_client_id"`
	AuthClientSecret string `json:"auth_client_secret"`

	MaxDistance  int    `json:"max_distance"`
	LatitudeRef  string `json:"latitude_ref"`
	LongitudeRef string `json:"longitude_ref"`

	UserServiceGRPC    string `json:"user_service_grpc"`
	ProductServiceGRPC string `json:"product_service_grpc"`
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

type PublisherName struct {
	ProductUpdateStock      string `json:"product_update_stock"`
	OrderPublishName        string `json:"order_publish_name"`
	EmailUpdateStatus       string `json:"email_update_status"`
	PublisherDeleteOrder    string `json:"publisher_delete_order"`
	PublisherPaymentSuccess string `json:"publisher_payment_success"`
	PublisherUpdateStatus   string `json:"publisher_update_status"`
	ProductToOrder          string `json:"produt_to_order"`
	UserLocationUpdate      string `json:"user_location_update"`
}

type ElasticSearch struct {
	Host string `json:"host"`
}

type Config struct {
	App           App           `json:"app"`
	Postgres      PostgresDB    `json:"postgres"`
	Redis         Redis         `json:"redis"`
	Publisher     PublisherName `json:"publisher"`
	RabbitMQ      RabbitMQ      `json:"rabbitmq"`
	ElasticSearch ElasticSearch `json:"elasticsearch"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort:              viper.GetString("APP_PORT"),
			AppEnv:               viper.GetString("APP_ENV"),
			JwtSecret:            viper.GetString("JWT_SECRET"),
			ServerTimeOut:        viper.GetInt("SERVER_TIMEOUT"),
			ApiGatewayServiceUrl: viper.GetString("APIGATEWAY_SERVICE_URL"),
			ProductServiceUrl:    viper.GetString("PRODUCT_SERVICE_URL"),
			UserServiceUrl:       viper.GetString("USER_SERVICE_URL"),
			AuthClientID:         viper.GetString("AUTH_CLIENT_ID"),
			AuthClientSecret:     viper.GetString("AUTH_CLIENT_SECRET"),
			LatitudeRef:          viper.GetString("LATITUDE_REF"),
			LongitudeRef:         viper.GetString("LONGITUDE_REF"),
			MaxDistance:          viper.GetInt("MAX_DISTANCE"),
			UserServiceGRPC:      viper.GetString("USER_SERVICE_GRPC"),
			ProductServiceGRPC:   viper.GetString("PRODUCT_SERVICE_GRPC"),
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
		Publisher: PublisherName{
			ProductUpdateStock:      viper.GetString("PRODUCT_UPDATE_STOCK"),
			OrderPublishName:        viper.GetString("ORDER_PUBLISH_NAME"),
			EmailUpdateStatus:       viper.GetString("EMAIL_UPDATE_STATUS"),
			PublisherDeleteOrder:    viper.GetString("PUBLISHER_DELETE_ORDER"),
			PublisherPaymentSuccess: viper.GetString("PUBLISHER_PAYMENT_SUCCESS"),
			PublisherUpdateStatus:   viper.GetString("PUBLISHER_UPDATE_STATUS"),
			ProductToOrder:          viper.GetString("PRODUCT_TO_ORDER"),
			UserLocationUpdate:      viper.GetString("USER_LOCATION_UPDATE"),
		},
		RabbitMQ: RabbitMQ{
			Host:        viper.GetString("RABBITMQ_HOST"),
			Port:        viper.GetString("RABBITMQ_PORT"),
			User:        viper.GetString("RABBITMQ_USER"),
			Password:    viper.GetString("RABBITMQ_PASSWORD"),
			VirtualHost: viper.GetString("RABBITMQ_VIRTUAL_HOST"),
		},
		ElasticSearch: ElasticSearch{
			Host: viper.GetString("ELASTICSEARCH_HOST"),
		},
	}
}
