package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/client"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/config"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/handlers"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/handlers/middleware"
	"github.com/swaggo/http-swagger/v2"

	_ "github.com/e1esm/casino-transaction-system/api-gateway/docs"
)

// @title Transaction Manager API
// @version 1.0

// @BasePath  /api/v1

// @contact.name Egor Mikhaylov
// @contact.email e.mikhaylov.dev@gmail.com
func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	cfg := mustParseConfig()
	cli := createTxManagerClient(cfg.Client)
	mx := createHttpHandler(cli)

	go runHttpServer(cfg.Http, mx)

	<-signalChan
}

func mustParseConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	return cfg
}

func runHttpServer(cfg config.HttpConfig, handler http.Handler) {
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler); err != nil {
		log.Fatalf("error starting http server: %v", err)
	}
}

func createHttpHandler(managerClient *client.TxManagerClient) http.Handler {
	mx := http.NewServeMux()
	h := handlers.New(managerClient)

	mx.HandleFunc("GET /api/v1/transactions/{id}", h.GetTransactionByID)
	mx.HandleFunc("GET /api/v1/transactions", h.GetTransactions)
	mx.HandleFunc("GET /ping", h.Healthcheck)

	mx.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return middleware.RecoveryMiddleware(mx)
}

func createTxManagerClient(clientConfig config.TxManagerClientConfig) *client.TxManagerClient {
	cli, err := client.New(clientConfig)
	if err != nil {
		log.Fatalf("error creating transaction manager kafka: %v", err)
	}

	return cli
}
