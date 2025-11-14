package types

import (
	"time"

	"github.com/google/uuid"
)

type FailedEntry struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
	Err   error  `json:"reason"`
}

type Transaction struct {
	UserID          uuid.UUID `json:"user_id" validate:"required,uuid"`
	TransactionType string    `json:"transaction_type" validate:"required,oneof=bet win"`
	Amount          int       `json:"amount" validate:"required,gt=0"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
}
