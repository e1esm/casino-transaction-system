package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
)

func writeJSONError(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": errMsg}); err != nil {
		log.Println("failed to write json response: ", err)
	}
}

func convertTransactionEntitiesToResponse(entities []entities.Transaction, total int) transactions {
	response := make([]transaction, 0, len(entities))
	for _, entity := range entities {
		response = append(response, convertTransactionEntityToResponse(entity))
	}

	return transactions{
		Transactions: response,
		Total:        total,
	}
}

func convertTransactionEntityToResponse(tr entities.Transaction) transaction {
	return transaction{
		ID:              tr.ID,
		UserID:          tr.UserID,
		Amount:          tr.Amount,
		TransactionType: string(tr.Type),
		TransactionDate: time.Unix(tr.Timestamp, 0).UTC(),
	}
}

func parseFiltersStruct(filters string) (entities.TransactionFilter, error) {
	if filters == "" {
		return entities.TransactionFilter{}, nil
	}

	resp := entities.TransactionFilter{}
	if err := json.Unmarshal([]byte(filters), &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func strToIntWithDefault(str string, def int64) int64 {
	i, err := strconv.Atoi(str)
	if err != nil {
		return def
	}

	return int64(i)
}
