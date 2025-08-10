package config

import "github.com/spf13/viper"

type App struct {
	AppPort   string `json:"app_port"`
	AppEnv    string `json:"app_env"`
	JwtSecret string `json:"jwt_secret"`
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

type Config struct {
	App      App        `json:"app"`
	Postgres PostgresDB `json:"postgres"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort:   viper.GetString("APP_PORT"),
			AppEnv:    viper.GetString("APP_ENV"),
			JwtSecret: viper.GetString("JWT_SECRET"),
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
	}
}