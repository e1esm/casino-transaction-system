package client

import (
	"errors"
	"fmt"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	txProto "github.com/e1esm/casino-transaction-system/api-gateway/src/internal/proto/tx-manager"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/svcerr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	transactionTypeProtoToEntity = map[txProto.TransactionType]entities.TransactionType{
		txProto.TransactionType_Bet: entities.Bet,
		txProto.TransactionType_Win: entities.Win,
	}

	transactionTypeEntityToProto = map[entities.TransactionType]txProto.TransactionType{
		entities.Bet: txProto.TransactionType_Bet,
		entities.Win: txProto.TransactionType_Win}
)

func mapReturnedCodeToSvcError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("internal error: %v", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %s", svcerr.ErrNotFound, st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", svcerr.ErrBadField, st.Message())
	default:
		return err
	}
}

func convertProtoTransactionsToEntities(transactions []*txProto.Transaction) ([]entities.Transaction, error) {
	resp := make([]entities.Transaction, 0, len(transactions))

	for _, t := range transactions {
		tmp, err := convertProtoTransactionToEntity(t)
		if err != nil {
			return nil, err
		}

		resp = append(resp, tmp)
	}

	return resp, nil
}

func convertProtoTransactionToEntity(transaction *txProto.Transaction) (entities.Transaction, error) {
	if transaction == nil {
		return entities.Transaction{}, errors.New("transaction is empty")
	}
	id, err := uuid.Parse(transaction.Id)
	if err != nil {
		return entities.Transaction{}, err
	}

	userID, err := uuid.Parse(transaction.UserId)
	if err != nil {
		return entities.Transaction{}, err
	}

	return entities.Transaction{
		ID:        id,
		UserID:    userID,
		Amount:    transaction.Amount,
		Timestamp: transaction.Timestamp,
		Type:      transactionTypeProtoToEntity[transaction.Type],
	}, nil
}
