package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/types"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Validator interface {
	Struct(t interface{}) error
}

type SaverService interface {
	Create(ctx context.Context, transactions ...models.Transaction) error
}

type DLQProducer interface {
	Produce(ctx context.Context, entries []types.FailedEntry)
}

type Client struct {
	client      *kgo.Client
	validator   Validator
	txSaver     SaverService
	dlqProducer DLQProducer

	maxRecordsPoll       int
	maxRetrySaveAttempts int
}

func NewWithClient(cli *kgo.Client, txSaver SaverService, validator Validator, producer DLQProducer, maxPolled, maxRetries int) *Client {
	return &Client{
		client:               cli,
		validator:            validator,
		txSaver:              txSaver,
		dlqProducer:          producer,
		maxRecordsPoll:       maxPolled,
		maxRetrySaveAttempts: maxRetries,
	}
}

func NewWithConfig(cfg config.KafkaConfig, txSaver SaverService, validator Validator, producer DLQProducer) (*Client, error) {
	if err := validate(cfg); err != nil {
		return nil, err
	}

	cli, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		kgo.ConsumerGroup(cfg.ConsumerConfig.ConsumerGroup),
		kgo.ConsumeTopics(cfg.ConsumerConfig.Topic),
		kgo.DisableAutoCommit(),
	)

	if err != nil {
		return nil, err
	}

	return NewWithClient(
		cli,
		txSaver,
		validator,
		producer,
		cfg.ConsumerConfig.MaxFetchedRecords,
		cfg.ConsumerConfig.MaxRetries,
	), nil
}

func (c *Client) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.client.Close()
			return nil
		default:
			failedEntries, err := c.consume(ctx)
			if err != nil {
				log.Println("consume failed: ", err)
				continue
			}

			if len(failedEntries) > 0 {
				c.dlqProducer.Produce(ctx, failedEntries)
			}
		}
	}
}

func (c *Client) consume(ctx context.Context) ([]types.FailedEntry, error) {
	fetches := c.client.PollRecords(ctx, c.maxRecordsPoll)

	failedEntries := make([]types.FailedEntry, 0)
	transactions := make([]models.Transaction, 0, len(fetches))

	fetches.EachRecord(func(r *kgo.Record) {
		var t types.Transaction

		err := json.Unmarshal(r.Value, &t)
		if err != nil {
			failedEntries = append(failedEntries, failedEntry(r, err))
			return
		}

		if err = c.validator.Struct(&t); err != nil {
			failedEntries = append(failedEntries, failedEntry(r, err))
			return
		}

		transactions = append(transactions, convertTransactionToModel(t))
	})

	if err := c.retry(func() error {
		return c.txSaver.Create(ctx, transactions...)
	}); err != nil {
		log.Println("Failed to insert transaction in the database: ", err.Error())
	}

	if err := c.client.CommitRecords(ctx, fetches.Records()...); err != nil {
		return nil, fmt.Errorf("faield to commit records: %w", err)
	}

	return failedEntries, nil
}

func (c *Client) retry(f func() error) error {
	var err error

	for i := range c.maxRetrySaveAttempts {
		err = f()
		if err == nil {
			return nil
		}

		if i == c.maxRetrySaveAttempts-1 {
			break
		}

		sleep := time.Second * time.Duration(1+i)
		jitter := time.Duration(rand.Intn(500)) * time.Millisecond
		time.Sleep(sleep + jitter)
	}

	return err
}

func failedEntry(r *kgo.Record, err error) types.FailedEntry {
	return types.FailedEntry{
		Key:   string(r.Key),
		Value: r.Value,
		Err:   err,
	}
}

func convertTransactionToModel(tx types.Transaction) models.Transaction {
	return models.Transaction{
		UserID:          tx.UserID,
		Type:            models.TransactionType(tx.TransactionType),
		Amount:          tx.Amount,
		TransactionTime: tx.TransactionDate,
	}
}

func validate(cfg config.KafkaConfig) error {
	if cfg.ConsumerConfig.MaxRetries == 0 {
		return fmt.Errorf("%w: max retries is zero", svcerr.ErrBadField)
	}

	if cfg.ConsumerConfig.MaxFetchedRecords == 0 {
		return fmt.Errorf("%w: max records is zero", svcerr.ErrBadField)
	}

	if len(cfg.Host) == 0 || (cfg.Port < 0 || cfg.Port > 65535) {
		return fmt.Errorf("%w: invalid host or port", svcerr.ErrBadField)
	}

	if len(cfg.ConsumerConfig.Topic) == 0 {
		return fmt.Errorf("%w: empty topic", svcerr.ErrBadField)
	}

	return nil
}
