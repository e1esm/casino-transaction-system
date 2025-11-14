package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/types"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Client struct {
	client *kgo.Client

	topic string
}

func NewClient(cfg config.KafkaConfig) (*Client, error) {
	cli, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		kgo.ConsumeTopics(cfg.ProducerConfig.Topic),
	)

	if err != nil {
		return nil, err
	}

	return &Client{
		client: cli,
		topic:  cfg.ProducerConfig.Topic,
	}, nil
}

func (c *Client) Produce(ctx context.Context, entries []types.FailedEntry) {
	for _, entry := range entries {

		resp, e := json.Marshal(entry)
		if e != nil {
			log.Printf("failed to marshal entry: %v", e)
			continue
		}

		c.client.Produce(ctx, &kgo.Record{
			Key:   []byte(entry.Key),
			Value: resp,
			Topic: c.topic,
		}, func(r *kgo.Record, err error) {
			if err != nil {
				log.Printf("failed to produce entry: %v", err)
			}
		})
	}
}

func (c *Client) Close() {
	c.client.Close()
}
