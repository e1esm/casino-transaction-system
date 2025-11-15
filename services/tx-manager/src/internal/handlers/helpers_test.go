package handlers

import (
	"testing"
	"time"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConvertTransactionModelToProto(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		in   models.Transaction
		want *proto.Transaction
	}{
		{
			name: "bet transaction",
			in: models.Transaction{
				ID:              id,
				UserID:          userID,
				Type:            models.Bet,
				Amount:          100,
				TransactionTime: now,
			},
			want: &proto.Transaction{
				Id:        id.String(),
				UserId:    userID.String(),
				Type:      proto.TransactionType_Bet,
				Amount:    100,
				Timestamp: now.Unix(),
			},
		},
		{
			name: "win transaction",
			in: models.Transaction{
				ID:              id,
				UserID:          userID,
				Type:            models.Win,
				Amount:          500,
				TransactionTime: now,
			},
			want: &proto.Transaction{
				Id:        id.String(),
				UserId:    userID.String(),
				Type:      proto.TransactionType_Win,
				Amount:    500,
				Timestamp: now.Unix(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTransactionModelToProto(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertTransactionsModelToProto(t *testing.T) {
	now := time.Now()

	id1, id2 := uuid.New(), uuid.New()
	user1, user2 := uuid.New(), uuid.New()

	tests := []struct {
		name  string
		input []models.Transaction
		want  []*proto.Transaction
	}{
		{
			name: "two transactions",
			input: []models.Transaction{
				{
					ID:              id1,
					UserID:          user1,
					Type:            models.Bet,
					Amount:          100,
					TransactionTime: now,
				},
				{
					ID:              id2,
					UserID:          user2,
					Type:            models.Win,
					Amount:          200,
					TransactionTime: now,
				},
			},
			want: []*proto.Transaction{
				{
					Id:        id1.String(),
					UserId:    user1.String(),
					Type:      proto.TransactionType_Bet,
					Amount:    100,
					Timestamp: now.Unix(),
				},
				{
					Id:        id2.String(),
					UserId:    user2.String(),
					Type:      proto.TransactionType_Win,
					Amount:    200,
					Timestamp: now.Unix(),
				},
			},
		},
		{
			name:  "empty slice",
			input: []models.Transaction{},
			want:  []*proto.Transaction{},
		},
		{
			name: "single transaction",
			input: []models.Transaction{
				{
					ID:              id1,
					UserID:          user1,
					Type:            models.Bet,
					Amount:          777,
					TransactionTime: now,
				},
			},
			want: []*proto.Transaction{
				{
					Id:        id1.String(),
					UserId:    user1.String(),
					Type:      proto.TransactionType_Bet,
					Amount:    777,
					Timestamp: now.Unix(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTransactionsModelToProto(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertProtoFiltersToModel(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		in        *proto.Filters
		want      models.TransactionFilter
		wantErr   bool
		errSubstr string
	}{
		{
			name: "nil input returns empty filter",
			in:   nil,
			want: models.TransactionFilter{},
		},
		{
			name: "valid user ID + type = Bet",
			in: &proto.Filters{
				UserId: userID.String(),
				Type:   proto.TransactionType_Bet,
			},
			want: models.TransactionFilter{
				UserID: &userID,
				Type:   ptr(models.Bet),
			},
		},
		{
			name: "valid user ID + type = Win",
			in: &proto.Filters{
				UserId: userID.String(),
				Type:   proto.TransactionType_Win,
			},
			want: models.TransactionFilter{
				UserID: &userID,
				Type:   ptr(models.Win),
			},
		},
		{
			name: "invalid user ID gives error",
			in: &proto.Filters{
				UserId: "NOT-A-UUID",
				Type:   proto.TransactionType_Bet,
			},
			wantErr:   true,
			errSubstr: "failed to parse user id",
		},
		{
			name: "valid user ID but unknown type → Type = nil",
			in: &proto.Filters{
				UserId: userID.String(),
				Type:   proto.TransactionType(999),
			},
			want: models.TransactionFilter{
				UserID: &userID,
				Type:   nil,
			},
		},
		{
			name: "only type set",
			in: &proto.Filters{
				Type: proto.TransactionType_Win,
			},
			want: models.TransactionFilter{
				UserID: nil,
				Type:   ptr(models.Win),
			},
		},
		{
			name: "only user ID set",
			in: &proto.Filters{
				UserId: userID.String(),
			},
			want: models.TransactionFilter{
				UserID: &userID,
				Type:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertProtoFiltersToModel(tt.in)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptr[T any](v T) *T { return &v }
