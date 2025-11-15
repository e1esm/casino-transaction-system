package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/types"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Client struct {
	client *kgo.Client

	topic string
}

func NewWithClient(client *kgo.Client, topic string) *Client {
	return &Client{
		client: client,
		topic:  topic,
	}
}

func NewWithConfig(cfg config.KafkaConfig) (*Client, error) {
	if err := validate(cfg); err != nil {
		return nil, err
	}

	cli, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		kgo.ConsumeTopics(cfg.ProducerConfig.Topic),
	)

	if err != nil {
		return nil, err
	}

	return NewWithClient(cli, cfg.ProducerConfig.Topic), nil
}

func (c *Client) Produce(ctx context.Context, entries []types.FailedEntry) {
	for _, entry := range entries {

		resp, err := json.Marshal(entry)
		if err != nil {
			log.Printf("failed to marshal entry: %v", err)
		}

		c.client.Produce(ctx, &kgo.Record{
			Key:   []byte(entry.Key),
			Value: resp,
			Topic: c.topic,
		}, nil)
	}
}

func (c *Client) Close() {
	c.client.Close()
}

func validate(cfg config.KafkaConfig) error {
	if len(cfg.Host) == 0 || (cfg.Port < 0 || cfg.Port > 65535) {
		return fmt.Errorf("%w: invalid host or port", svcerr.ErrBadField)
	}

	if len(cfg.ProducerConfig.Topic) == 0 {
		return fmt.Errorf("%w: empty topic", svcerr.ErrBadField)
	}

	return nil
}
