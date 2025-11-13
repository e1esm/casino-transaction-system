package config

import "github.com/caarlos0/env/v11"

type TxManagerClientConfig struct {
	Host string `env:"HOST,required"`
}

type HttpConfig struct {
	Port int64 `env:"PORT,required"`
}

type Config struct {
	Client TxManagerClientConfig `envPrefix:"TX_MANAGER_"`
	Http   HttpConfig            `envPrefix:"HTTP_"`
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
