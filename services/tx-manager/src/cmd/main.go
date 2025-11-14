package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/kafka/consumer"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/broker/kafka/dlq"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers/interceptors"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"
	txRepo "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/repository/transaction"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/service/transaction"
	"github.com/go-playground/validator/v10"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM)

	cfg := mustInitConfig()

	repo := mustInitRepository(cfg)
	txSvc := transaction.New(repo)
	dlqProducer := mustInitDLQProducer(cfg)
	broker := mustInitBroker(cfg, txSvc, dlqProducer)
	h := handlers.New(txSvc)
	srv := newGrpcServer(h)

	go serveGrpc(srv, cfg.Grpc)
	go broker.Consume(ctx)

	<-ctx.Done()
	cancelFunc()

	srv.GracefulStop()
	repo.Close()
}

func mustInitConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to load config: %v", err))
	}

	return cfg
}

func mustInitRepository(cfg *config.Config) *txRepo.Repository {
	repo, err := txRepo.New(cfg.Database)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to initialize repository: %v", err))
	}

	return repo
}

func mustInitDLQProducer(cfg *config.Config) *dlq.Client {
	cli, err := dlq.NewClient(cfg.Kafka)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to initialize DLQ client: %v", err))
	}

	return cli
}

func mustInitBroker(cfg *config.Config, txSvc *transaction.Service, dlqCli *dlq.Client) *consumer.Client {
	cli, err := consumer.New(cfg.Kafka, txSvc, validator.New(), dlqCli)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to initialize broker: %v", err))
	}

	return cli
}

func newGrpcServer(h *handlers.Handler) *grpc.Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RecoveryUnaryInterceptor),
	)

	proto.RegisterTransactionManagerServer(srv, h)

	return srv
}

func serveGrpc(srv *grpc.Server, grpcConfig config.GrpcConfig) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcConfig.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := srv.Serve(list); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
