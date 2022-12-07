package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	HttpPort      string
	Postgres      PostgresConfig
	Redis         Redis
	AuthSecretKey string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type Redis struct {
	Addr string
}

func Load(path string) Config {
	err := godotenv.Load(path + "/.env") // load .env file if it exists
	if err != nil {
		fmt.Printf("Error loading .env file: %v", err)
	}

	conf := viper.New()
	conf.AutomaticEnv()

	cfg := Config{
		HttpPort: conf.GetString("HTTP_PORT"),
		Postgres: PostgresConfig{
			Host:     conf.GetString("POSTGRES_HOST"),
			Port:     conf.GetString("POSTGRES_PORT"),
			User:     conf.GetString("POSTGRES_USER"),
			Password: conf.GetString("POSTGRES_PASSWORD"),
			Database: conf.GetString("POSTGRES_DATABASE"),
		},
		Redis: Redis{
			Addr: conf.GetString("REDIS_ADDR"),
		},
		AuthSecretKey: conf.GetString("AUTH_SECRET_KEY"),
	}

	return cfg
}
