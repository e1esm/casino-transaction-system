package handlers

import (
	"time"

	"github.com/google/uuid"
)

type transaction struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Amount          int64     `json:"amount"`
	TransactionType string    `json:"type"`
	TransactionDate time.Time `json:"date"`
}

type transactions struct {
	Transactions []transaction `json:"transactions"`
	Total        int           `json:"total"`
}
