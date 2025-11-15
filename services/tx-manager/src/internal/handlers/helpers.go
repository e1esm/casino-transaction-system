package handlers

import (
	"fmt"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"
	"github.com/google/uuid"
)

var txTypeProtoToModel = map[proto.TransactionType]models.TransactionType{
	proto.TransactionType_Bet: models.Bet,
	proto.TransactionType_Win: models.Win,
}

var txTypeModelToProto = map[models.TransactionType]proto.TransactionType{
	models.Bet: proto.TransactionType_Bet,
	models.Win: proto.TransactionType_Win,
}

func convertTransactionsModelToProto(transactions []models.Transaction) []*proto.Transaction {
	protoTransactions := make([]*proto.Transaction, 0, len(transactions))

	for _, transaction := range transactions {
		protoTransactions = append(protoTransactions, convertTransactionModelToProto(transaction))
	}

	return protoTransactions
}

func convertTransactionModelToProto(tr models.Transaction) *proto.Transaction {
	return &proto.Transaction{
		Id:        tr.ID.String(),
		UserId:    tr.UserID.String(),
		Type:      txTypeModelToProto[tr.Type],
		Amount:    int64(tr.Amount),
		Timestamp: tr.TransactionTime.Unix(),
	}
}

func convertProtoFiltersToModel(req *proto.Filters) (models.TransactionFilter, error) {
	if req == nil {
		return models.TransactionFilter{}, nil
	}

	var (
		id     *uuid.UUID
		txType *models.TransactionType
	)

	if len(req.UserId) > 0 {
		parsedID, err := uuid.Parse(req.UserId)
		if err != nil {
			return models.TransactionFilter{}, fmt.Errorf("failed to parse user id: %w", err)
		}

		id = &parsedID
	}

	if v, ok := txTypeProtoToModel[req.Type]; ok {
		txType = &v
	}

	return models.TransactionFilter{
		UserID: id,
		Type:   txType,
	}, nil
}
