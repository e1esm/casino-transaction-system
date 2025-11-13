package config

import "github.com/caarlos0/env/v11"

type KafkaConfig struct {
	Host          string `env:"HOST,required"`
	Port          int    `env:"PORT" envDefault:"9092"`
	Topic         string `env:"TOPIC,required"`
	ConsumerGroup string `env:"CONSUMER_GROUP,required"`
}

type DatabaseConfig struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT" envDefault:"5432"`
	Username string `env:"USERNAME,required"`
	Password string `env:"PASSWORD,required"`
	Name     string `env:"NAME,required"`
	SSLMode  string `env:"SSL_MODE"`
}

type GrpcConfig struct {
	Port int `env:"PORT,required"`
}

type Config struct {
	Kafka    KafkaConfig    `envPrefix:"BROKER_"`
	Database DatabaseConfig `envPrefix:"DATABASE_"`
	Grpc     GrpcConfig     `envPrefix:"GRPC_"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
