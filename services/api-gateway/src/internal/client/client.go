package client

import (
	"context"
	"time"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/config"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	txProto "github.com/e1esm/casino-transaction-system/api-gateway/src/internal/proto/tx-manager"
	"google.golang.org/grpc/codes"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TxManagerClient struct {
	cli txProto.TransactionManagerClient
}

func New(config config.TxManagerClientConfig) (*TxManagerClient, error) {
	cli, err := grpc.NewClient(config.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(
			retry.WithMax(10),
			retry.WithCodes(codes.Unavailable),
			retry.WithPerRetryTimeout(time.Second),
			retry.WithBackoff(retry.BackoffLinear(500*time.Millisecond)),
		)),
	)
	if err != nil {
		return nil, err
	}

	return &TxManagerClient{
		txProto.NewTransactionManagerClient(cli),
	}, nil
}

func (c *TxManagerClient) GetTransactionByID(ctx context.Context, id uuid.UUID) (entities.Transaction, error) {
	resp, err := c.cli.GetTransactionByID(ctx, &txProto.GetTransactionByIDRequest{
		Id: id.String(),
	})
	if err != nil {
		return entities.Transaction{}, mapReturnedCodeToSvcError(err)
	}

	tx, err := convertProtoTransactionToEntity(resp.Transaction)
	if err != nil {
		return entities.Transaction{}, err
	}

	return tx, nil
}

func (c *TxManagerClient) GetTransactions(
	ctx context.Context,
	filter entities.TransactionFilter,
	orderBy string,
	limit,
	offset int64) ([]entities.Transaction, int, error) {

	resp, err := c.cli.GetTransactionByFilters(ctx, &txProto.GetTransactionByFiltersRequest{
		Filters: &txProto.Filters{
			Type:   transactionTypeEntityToProto[filter.Type],
			UserId: filter.UserID,
		},
		OrderBy: orderBy,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, mapReturnedCodeToSvcError(err)
	}

	transactions, err := convertProtoTransactionsToEntities(resp.Transaction)
	if err != nil {
		return nil, 0, err
	}

	return transactions, len(transactions), nil
}
