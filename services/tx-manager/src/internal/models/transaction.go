package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

var (
	Bet TransactionType = "bet"
	Win TransactionType = "win"
)

type Transaction struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Type            TransactionType
	Amount          int
	TransactionTime time.Time
}

type TransactionFilter struct {
	UserID *uuid.UUID
	Type   *TransactionType
}

func (tf TransactionFilter) String() (string, []any) {
	var conditions []string
	var args []any

	argPos := 1

	if tf.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, *tf.UserID)
		argPos++
	}

	if tf.Type != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_type = $%d", argPos))
		args = append(args, *tf.Type)
		argPos++
	}

	if len(conditions) > 0 {
		return strings.Join(conditions, " AND "), args
	}

	return "", nil
}
