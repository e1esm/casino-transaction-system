package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers/interceptors"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"
	txRepo "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/repository/transaction"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/service/transaction"

	"google.golang.org/grpc"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	cfg := mustInitConfig()
	repo := mustInitRepository(cfg)
	txSvc := transaction.New(repo)
	h := handlers.New(txSvc)
	srv := newGrpcServer(h)

	go serveGrpc(srv, cfg.Grpc)

	<-signalChan

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
