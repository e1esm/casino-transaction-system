package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/kafka/consumer/mocks"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/types"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	net "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	kafkaHost, kafkaPort string
	testTopic            = "test-transaction"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	network, err := net.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer network.Remove(ctx)

	zkReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-zookeeper:7.4.0",
		ExposedPorts: []string{"2181/tcp"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
		},
		Networks:   []string{network.Name},
		Hostname:   "zookeeper",
		WaitingFor: wait.ForListeningPort("2181/tcp").WithStartupTimeout(60 * time.Second),
	}

	zkC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: zkReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start zookeeper: %v", err)
	}
	defer zkC.Terminate(ctx)

	kafkaReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.4.0",
		ExposedPorts: []string{"9094/tcp"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
			"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:9094",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://127.0.0.1:9094",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "PLAINTEXT:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":       "PLAINTEXT",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":        "true",
		},
		Networks:   []string{network.Name},
		Hostname:   "kafka",
		WaitingFor: wait.ForLog("started (kafka.server.KafkaServer)").WithStartupTimeout(120 * time.Second),
		HostConfigModifier: func(config *container.HostConfig) {
			config.PortBindings = nat.PortMap{
				"9094/tcp": []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "9094"},
				},
			}
		},
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start kafka: %v", err)
	}
	defer kafkaC.Terminate(ctx)

	kafkaHost, _ = kafkaC.Host(ctx)
	nPort, _ := kafkaC.MappedPort(ctx, "9094")
	kafkaPort = nPort.Port()

	kafkaC.Exec(ctx, []string{
		"kafka-topics",
		"--create",
		"--topic", testTopic,
		"--bootstrap-server", fmt.Sprintf("%s:%s", kafkaHost, kafkaPort),
		"--partitions", "1",
		"--replication-factor", "1",
	})

	os.Exit(m.Run())
}

func TestClientRetry(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		failTimes     int
		expectErr     bool
		expectedCalls int
	}{
		{
			name:          "succeeds first attempt",
			maxRetries:    3,
			failTimes:     0,
			expectErr:     false,
			expectedCalls: 1,
		},
		{
			name:          "succeeds after retries",
			maxRetries:    3,
			failTimes:     2,
			expectErr:     false,
			expectedCalls: 3,
		},
		{
			name:          "fails after all retries",
			maxRetries:    3,
			failTimes:     5,
			expectErr:     true,
			expectedCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				maxRetrySaveAttempts: tt.maxRetries,
			}

			attempts := 0
			err := c.retry(func() error {
				attempts++
				if attempts <= tt.failTimes {
					return errors.New("fail")
				}
				return nil
			})

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCalls, attempts)

			if attempts != tt.expectedCalls {
				t.Errorf("expected %d calls, got %d", tt.expectedCalls, attempts)
			}
		})
	}
}

func TestFailedEntry(t *testing.T) {
	tests := []struct {
		name   string
		record *kgo.Record
		err    error
		want   types.FailedEntry
	}{
		{
			name: "normal record",
			record: &kgo.Record{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
			err: errors.New("some error"),
			want: types.FailedEntry{
				Key:   "key1",
				Value: []byte("value1"),
				Err:   errors.New("some error"),
			},
		},
		{
			name: "nil error",
			record: &kgo.Record{
				Key:   []byte("key2"),
				Value: []byte("value2"),
			},
			err: nil,
			want: types.FailedEntry{
				Key:   "key2",
				Value: []byte("value2"),
				Err:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &kgo.Record{
				Key:   tt.record.Key,
				Value: tt.record.Value,
			}
			got := failedEntry(r, tt.err)

			assert.Equal(t, tt.want.Key, got.Key)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.Err, got.Err)
		})
	}
}

func TestConvertTransactionToModel(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	tests := []struct {
		name string
		tx   types.Transaction
		want models.Transaction
	}{
		{
			name: "basic conversion",
			tx: types.Transaction{
				UserID:          id,
				TransactionType: "bet",
				Amount:          100,
				TransactionDate: now,
			},
			want: models.Transaction{
				UserID:          id,
				Type:            models.TransactionType("bet"),
				Amount:          100,
				TransactionTime: now,
			},
		},
		{
			name: "zero values",
			tx:   types.Transaction{},
			want: models.Transaction{
				UserID:          uuid.Nil,
				Type:            models.TransactionType(""),
				Amount:          0,
				TransactionTime: time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTransactionToModel(tt.tx)

			assert.Equal(t, tt.want.UserID, got.UserID)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.Amount, got.Amount)
			assert.Equal(t, tt.want.TransactionTime, got.TransactionTime)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.KafkaConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        3,
					MaxFetchedRecords: 10,
					Topic:             "test-topic",
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic",
				},
			},
			wantErr: false,
		},
		{
			name: "max retries zero",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        0,
					MaxFetchedRecords: 10,
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic",
				},
			},
			wantErr: true,
			errMsg:  "max retries is zero",
		},
		{
			name: "max fetched records zero",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        3,
					MaxFetchedRecords: 0,
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic",
				},
			},
			wantErr: true,
			errMsg:  "max records is zero",
		},
		{
			name: "invalid host",
			cfg: config.KafkaConfig{
				Host: "",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        3,
					MaxFetchedRecords: 10,
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic",
				},
			},
			wantErr: true,
			errMsg:  "invalid host or port",
		},
		{
			name: "invalid port",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 70000,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        3,
					MaxFetchedRecords: 10,
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic",
				},
			},
			wantErr: true,
			errMsg:  "invalid host or port",
		},
		{
			name: "empty topic",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					MaxRetries:        3,
					MaxFetchedRecords: 10,
				},
				ProducerConfig: config.ProducerConfig{
					Topic: "",
				},
			},
			wantErr: true,
			errMsg:  "empty topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errMsg)
				assert.ErrorIs(t, err, svcerr.ErrBadField)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestClient_Consume_Integration(t *testing.T) {
	tests := []struct {
		name          string
		messages      []interface{}
		expectedTx    int
		expectedDLQ   int
		validationErr error
		saverErr      error
	}{
		{
			name: "all valid messages",
			messages: []interface{}{
				types.Transaction{UserID: uuid.New(), TransactionType: "bet", Amount: 100, TransactionDate: time.Now()},
				types.Transaction{UserID: uuid.New(), TransactionType: "win", Amount: 50, TransactionDate: time.Now()},
			},
			expectedTx:  1,
			expectedDLQ: 0,
		},
		{
			name: "one invalid message",
			messages: []interface{}{
				types.Transaction{UserID: uuid.New(), TransactionType: "bet", Amount: 100, TransactionDate: time.Now()},
				"{ \"UserID\": \"u1\", \"Amount\": 100, \"TransactionType\": \"bet\"",
			},
			validationErr: &validator.InvalidValidationError{},
			expectedTx:    1,
			expectedDLQ:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			v := mocks.NewMockValidator(t)
			saver := mocks.NewMockSaverService(t)
			dlq := mocks.NewMockDLQProducer(t)
			kCli := newKafkaClient()

			v.On("Struct", mock.Anything).Return(tt.validationErr)
			saver.On("Create", mock.Anything, mock.Anything).Return(tt.saverErr)
			if tt.expectedDLQ > 0 {
				dlq.On("Produce", mock.Anything, mock.Anything)
			}

			c := NewWithClient(kCli, saver, v, dlq, 10, 10)

			produceMessages(t, kCli, testTopic, tt.messages...)

			go func() {
				_ = c.Consume(ctx)
			}()

			time.Sleep(5 * time.Second)
			cancel()

			assert.Len(t, saver.Calls, tt.expectedTx)
			assert.Len(t, dlq.Calls, tt.expectedDLQ)
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name          string
		cfg           config.KafkaConfig
		mockValidator Validator
		mockSaver     SaverService
		mockDLQ       DLQProducer
		expectErr     bool
		errMsg        string
	}{
		{
			name: "valid config",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					ConsumerGroup:     "cg",
					Topic:             "test-topic",
					MaxFetchedRecords: 10,
					MaxRetries:        3,
				},
			},
			mockValidator: mocks.NewMockValidator(t),
			mockSaver:     mocks.NewMockSaverService(t),
			mockDLQ:       mocks.NewMockDLQProducer(t),
			expectErr:     false,
		},
		{
			name: "invalid config (max retries zero)",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					ConsumerGroup:     "cg",
					Topic:             "test-topic",
					MaxFetchedRecords: 10,
					MaxRetries:        0,
				},
			},
			mockValidator: mocks.NewMockValidator(t),
			mockSaver:     mocks.NewMockSaverService(t),
			mockDLQ:       mocks.NewMockDLQProducer(t),
			expectErr:     true,
			errMsg:        "max retries is zero",
		},
		{
			name: "invalid config (empty topic)",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ConsumerConfig: config.ConsumerConfig{
					ConsumerGroup:     "cg",
					Topic:             "",
					MaxFetchedRecords: 10,
					MaxRetries:        3,
				},
			},
			mockValidator: mocks.NewMockValidator(t),
			mockSaver:     mocks.NewMockSaverService(t),
			mockDLQ:       mocks.NewMockDLQProducer(t),
			expectErr:     true,
			errMsg:        "empty topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewWithConfig(tt.cfg, tt.mockSaver, tt.mockValidator, tt.mockDLQ)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func produceMessages(t *testing.T, client *kgo.Client, topic string, messages ...interface{}) {
	records := make([]*kgo.Record, len(messages))
	for i, msg := range messages {
		b, err := json.Marshal(msg)
		assert.NoError(t, err)
		records[i] = &kgo.Record{
			Topic: topic,
			Value: b,
		}
	}
	resp := client.ProduceSync(context.Background(), records...)
	assert.NoError(t, resp.FirstErr())
}

func newKafkaClient() *kgo.Client {
	resp, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%s", kafkaHost, kafkaPort)),
		kgo.ConsumeTopics(testTopic),
		kgo.ConsumerGroup("test-gr"),
		kgo.DisableAutoCommit(),
	)

	if err != nil {
		log.Fatal(err)
	}

	return resp
}
