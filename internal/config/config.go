package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Dsn      string `env:"POSTGRES_DSN" env-required:"true"`
	HostPort string `env:"HOST_PORT" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
