package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Hash(t *testing.T) {
	now := time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)

	uid1 := uuid.New()
	uid2 := uuid.New()

	tests := []struct {
		name         string
		tx           Transaction
		tx2          Transaction
		expectedSame bool
	}{
		{
			name: "same transactions produce same hash",
			tx: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			tx2: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			expectedSame: true,
		},
		{
			name: "different userID produces different hash",
			tx: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			tx2: Transaction{
				UserID:          uid2,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			expectedSame: false,
		},
		{
			name: "different type produces different hash",
			tx: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			tx2: Transaction{
				UserID:          uid1,
				Type:            Win,
				Amount:          100,
				TransactionTime: now,
			},
			expectedSame: false,
		},
		{
			name: "different amount produces different hash",
			tx: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			tx2: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          200,
				TransactionTime: now,
			},
			expectedSame: false,
		},
		{
			name: "different timestamp produces different hash",
			tx: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now,
			},
			tx2: Transaction{
				UserID:          uid1,
				Type:            Bet,
				Amount:          100,
				TransactionTime: now.Add(time.Minute),
			},
			expectedSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := tt.tx.Hash()
			h2 := tt.tx2.Hash()

			if tt.expectedSame {
				assert.Equal(t, h1, h2)
			} else {
				assert.NotEqual(t, h1, h2)
			}
		})
	}
}

func TestTransactionFilter_String(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()

	tests := []struct {
		name         string
		filter       TransactionFilter
		expectedSQL  string
		expectedArgs []any
	}{
		{
			name: "both UserID and Type set",
			filter: TransactionFilter{
				UserID: &userID1,
				Type:   ptrTransactionType(Bet),
			},
			expectedSQL:  "user_id = $1 AND transaction_type = $2",
			expectedArgs: []any{userID1, Bet},
		},
		{
			name: "only UserID set",
			filter: TransactionFilter{
				UserID: &userID2,
			},
			expectedSQL:  "user_id = $1",
			expectedArgs: []any{userID2},
		},
		{
			name: "only Type set",
			filter: TransactionFilter{
				Type: ptrTransactionType(Win),
			},
			expectedSQL:  "transaction_type = $1",
			expectedArgs: []any{Win},
		},
		{
			name:         "neither field set",
			filter:       TransactionFilter{},
			expectedSQL:  "",
			expectedArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := tt.filter.String()
			assert.Equal(t, tt.expectedSQL, gotSQL)
			assert.Equal(t, tt.expectedArgs, gotArgs)
		})
	}
}

func ptrTransactionType(t TransactionType) *TransactionType {
	return &t
}
