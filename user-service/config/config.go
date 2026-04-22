package config

import "github.com/spf13/viper"

type App struct {
	AppPort string `json:"app_port"`
	AppEnv  string `json:"app_env"`

	JwtSecret string `json:"jwt_secret"`
	JwtIssuer string `json:"jwt_issuer"`

	UrlFrontend string `json:"url_frontend"`

	AuthClientID     string `json:"auth_client_id"`
	AuthClientSecret string `json:"auth_client_secret"`
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

type Supabase struct {
	URL    string `json:"url"`
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
}

type Oauth struct {
	GoogleOauthClientID     string `json:"google_oauth_client_id"`
	GoogleOauthClientSecret string `json:"google_oauth_client_secret"`
	GoogleRedirectUrl       string `json:"google_redirect_url"`
}

type PublisherName struct {
	UserLocationUpdate string `json:"user_location_update"`
}

type Config struct {
	App           App           `json:"app"`
	Postgres      PostgresDB    `json:"postgres"`
	Redis         Redis         `json:"redis"`
	RabbitMQ      RabbitMQ      `json:"rabbitmq"`
	Storage       Supabase      `json:"supabase"`
	Oauth         Oauth         `json:"oauth"`
	PublisherName PublisherName `json:"publisher"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort:          viper.GetString("APP_PORT"),
			AppEnv:           viper.GetString("APP_ENV"),
			JwtSecret:        viper.GetString("JWT_SECRET"),
			JwtIssuer:        viper.GetString("JWT_ISSUER"),
			UrlFrontend:      viper.GetString("URL_FRONTEND"),
			AuthClientID:     viper.GetString("AUTH_CLIENT_ID"),
			AuthClientSecret: viper.GetString("AUTH_CLIENT_SECRET"),
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
		Storage: Supabase{
			URL:    viper.GetString("SUPABASE_STORAGE_URL"),
			Key:    viper.GetString("SUPABASE_STORAGE_KEY"),
			Bucket: viper.GetString("SUPABASE_STORAGE_BUCKET"),
		},
		Oauth: Oauth{
			GoogleOauthClientID:     viper.GetString("GOOGLE_OAUTH_CLIENT_ID"),
			GoogleOauthClientSecret: viper.GetString("GOOGLE_OAUTH_CLIENT_SECRET"),
			GoogleRedirectUrl:       viper.GetString("GOOGLE_REDIRECT_URL"),
		},
		PublisherName: PublisherName{
			UserLocationUpdate: viper.GetString("USER_LOCATION_UPDATE"),
		},
	}
}
