package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/types"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	net "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	cli       *kgo.Client
	testTopic = "test-dlq"
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
		ExposedPorts: []string{"9093/tcp"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
			"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:9093",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://127.0.0.1:9093",
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
				"9093/tcp": []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "9093"},
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

	kafkaHost, _ := kafkaC.Host(ctx)
	kafkaPort, _ := kafkaC.MappedPort(ctx, "9093")

	kafkaC.Exec(ctx, []string{
		"kafka-topics",
		"--create",
		"--topic", testTopic,
		"--bootstrap-server", fmt.Sprintf("%s:%s", kafkaHost, kafkaPort.Port()),
		"--partitions", "1",
		"--replication-factor", "1",
	})

	kCli, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%s", kafkaHost, kafkaPort.Port())),
		kgo.ConsumeTopics(testTopic),
		kgo.ConsumerGroup("test-gr"),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		log.Fatalf("failed to create kafka client: %v", err)
	}

	cli = kCli
	defer cli.Close()

	os.Exit(m.Run())
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.KafkaConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic-1",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid host",
			cfg: config.KafkaConfig{
				ProducerConfig: config.ProducerConfig{
					Topic: "test-topic-2",
				},
			},
			wantErr: true,
		},
		{
			name: "empty topic",
			cfg: config.KafkaConfig{
				Host: "127.0.0.1",
				Port: 9092,
				ProducerConfig: config.ProducerConfig{
					Topic: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewWithConfig(tt.cfg)
			if client != nil {
				defer client.Close()
			}

			assert.Equal(t, tt.wantErr, err != nil)

		})
	}
}

func TestProducerIntegration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		entries []types.FailedEntry
		wantN   int
	}{
		{
			name: "single message",
			entries: []types.FailedEntry{
				{Key: "k1", Value: []byte("v1")},
			},
			wantN: 1,
		},
		{
			name: "multiple messages",
			entries: []types.FailedEntry{
				{Key: "kA", Value: []byte("123")},
				{Key: "kB", Value: []byte("456")},
			},
			wantN: 2,
		},
		{
			name:    "empty batch",
			entries: nil,
			wantN:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewWithClient(cli, testTopic)
			client.Produce(ctx, tt.entries)

			msgs := consumeN(ctx, cli, tt.wantN, 5*time.Second)

			assert.Len(t, msgs, tt.wantN)

			for i, m := range msgs {
				if tt.wantN == 0 {
					continue
				}

				var got types.FailedEntry
				err := json.Unmarshal(m.Value, &got)
				assert.NoError(t, err)

				assert.Equal(t, tt.entries[i].Key, string(m.Key))
				assert.Equal(t, tt.entries[i].Value, got.Value)
			}
		})
	}
}

func consumeN(ctx context.Context, client *kgo.Client, n int, timeout time.Duration) []*kgo.Record {
	deadline := time.Now().Add(timeout)
	result := make([]*kgo.Record, 0, n)

	for len(result) < n && time.Now().Before(deadline) {
		fetches := client.PollFetches(ctx)

		cli.CommitRecords(ctx, fetches.Records()...)
		for _, fetch := range fetches.Records() {
			result = append(result, &kgo.Record{
				Key:       fetch.Key,
				Value:     fetch.Value,
				Timestamp: fetch.Timestamp,
			})

			if len(result) == n {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	return result
}
