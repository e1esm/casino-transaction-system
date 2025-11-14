package config

import "github.com/caarlos0/env/v11"

type ConsumerConfig struct {
	Topic             string `env:"TOPIC,required"`
	ConsumerGroup     string `env:"GROUP,required"`
	MaxFetchedRecords int    `env:"MAX_RECORDS_FETCHED,required"`
	MaxRetries        int    `env:"MAX_RETRIES,required"`
}

type ProducerConfig struct {
	Topic string `env:"TOPIC,required"`
}

type KafkaConfig struct {
	Host           string         `env:"HOST,required"`
	Port           int            `env:"PORT" envDefault:"9092"`
	ConsumerConfig ConsumerConfig `envPrefix:"CONSUMER_"`
	ProducerConfig ProducerConfig `envPrefix:"PRODUCER_"`
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
