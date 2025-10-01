package config

type App struct {
	AppPort   string
	AppEnv    string
	JwtSecret string
}

type PostgresDB struct {
	Host      string
	Port      string
	User      string
	Password  string
	DBName    string
	DBMaxOpen int
	DBMaxIdle int
}

type Redis struct {
	Host string
	Port string
}

type RabbitMQ struct {
	Host        string
	User        string
	Password    string
	VirtualHost string
	Port        string
}

type Email struct {
	Host string
	Port int
	User string
	Pass string
	Tls  bool
	From string
}

type Config struct {
	App      App
	Postgres PostgresDB
	RabbitMQ RabbitMQ
	Email    Email
	Redis    Redis
}

func NewConfig() *Config {
	env := LoadEnv()

	return &Config{
		App: App{
			AppPort:   env.AppPort,
			AppEnv:    env.AppEnv,
			JwtSecret: env.JwtSecret,
		},
		Postgres: PostgresDB{
			Host:      env.DatabaseHost,
			Port:      env.DatabasePort,
			User:      env.DatabaseUser,
			Password:  env.DatabasePassword,
			DBName:    env.DatabaseName,
			DBMaxOpen: env.DatabaseMaxOpen,
			DBMaxIdle: env.DatabaseMaxIdle,
		},
		Redis: Redis{
			Host: env.RedisHost,
			Port: env.RedisPort,
		},
		RabbitMQ: RabbitMQ{
			Host:        env.RabbitHost,
			Port:        env.RabbitPort,
			User:        env.RabbitUser,
			Password:    env.RabbitPassword,
			VirtualHost: env.RabbitVirtualHost,
		},
		Email: Email{
			Host: env.EmailHost,
			Port: env.EmailPort,
			User: env.EmailUser,
			Pass: env.EmailPass,
			Tls:  env.EmailTls,
			From: env.EmailFrom,
		},
	}
}
