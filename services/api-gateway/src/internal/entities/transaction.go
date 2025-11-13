package entities

import (
	"github.com/google/uuid"
)

type TransactionType string

var (
	Bet TransactionType = "bet"
	Win TransactionType = "win"
)

type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      TransactionType
	Amount    int64
	Timestamp int64
}

type TransactionFilter struct {
	UserID string
	Type   TransactionType
}
