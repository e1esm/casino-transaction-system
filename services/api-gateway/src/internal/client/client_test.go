package client

import (
	"context"
	"testing"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/client/mocks"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	txProto "github.com/e1esm/casino-transaction-system/api-gateway/src/internal/proto/tx-manager"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTxManagerClientGetTransactionByID(t *testing.T) {
	mockCli := new(mocks.MockProtoClient)
	client := NewClientFromProto(mockCli)
	id := uuid.New()

	tests := []struct {
		name        string
		reqID       uuid.UUID
		mockResp    *txProto.GetTransactionByIDResponse
		mockErr     error
		expectedErr bool
		expectedID  uuid.UUID
	}{
		{
			name:        "success",
			reqID:       id,
			mockResp:    &txProto.GetTransactionByIDResponse{Transaction: &txProto.Transaction{Id: id.String(), Amount: 100, UserId: uuid.New().String()}},
			mockErr:     nil,
			expectedErr: false,
			expectedID:  id,
		},
		{
			name:        "transaction not found",
			reqID:       uuid.New(),
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "transaction not found"),
			expectedErr: true,
			expectedID:  uuid.Nil,
		},
		{
			name:        "invalid transaction field: user id",
			reqID:       id,
			mockResp:    &txProto.GetTransactionByIDResponse{Transaction: &txProto.Transaction{Id: id.String(), Amount: 100, UserId: "invalid-id"}},
			mockErr:     nil,
			expectedErr: true,
			expectedID:  id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCli.On("GetTransactionByID",
				mock.Anything,
				mock.Anything,
			).Return(tt.mockResp, tt.mockErr).Once()

			tx, err := client.GetTransactionByID(context.Background(), id)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID.String(), tx.ID.String())
			}

			mockCli.AssertExpectations(t)
		})
	}
}

func TestTxManagerClient_GetTransactions(t *testing.T) {
	mockCli := new(mocks.MockProtoClient)
	client := NewClientFromProto(mockCli)

	userID := uuid.New().String()

	tests := []struct {
		name          string
		filter        entities.TransactionFilter
		orderBy       string
		limit         int64
		offset        int64
		mockResp      *txProto.GetTransactionByFiltersResponse
		mockErr       error
		expectedErr   bool
		expectedCount int
	}{
		{
			name: "success",
			filter: entities.TransactionFilter{
				Type:   entities.Bet,
				UserID: userID,
			},
			orderBy: "amount",
			limit:   10,
			offset:  0,
			mockResp: &txProto.GetTransactionByFiltersResponse{
				Transaction: []*txProto.Transaction{
					{Id: uuid.New().String(), UserId: uuid.NewString(), Amount: 100, Type: txProto.TransactionType_Bet},
					{Id: uuid.New().String(), UserId: uuid.NewString(), Amount: 200, Type: txProto.TransactionType_Bet},
				},
			},
			mockErr:       nil,
			expectedErr:   false,
			expectedCount: 2,
		},
		{
			name: "grpc error",
			filter: entities.TransactionFilter{
				Type:   entities.Win,
				UserID: userID,
			},
			orderBy:       "amount",
			limit:         5,
			offset:        0,
			mockResp:      nil,
			mockErr:       status.Error(codes.Unavailable, "unavailable"),
			expectedErr:   true,
			expectedCount: 0,
		},
		{
			name: "conversion error (invalid proto payload)",
			filter: entities.TransactionFilter{
				Type:   entities.Win,
				UserID: userID,
			},
			orderBy: "amount",
			limit:   5,
			offset:  0,
			mockResp: &txProto.GetTransactionByFiltersResponse{
				Transaction: []*txProto.Transaction{
					{Id: "invalid-uuid", Amount: 100},
				},
			},
			mockErr:       nil,
			expectedErr:   true,
			expectedCount: 0,
		},
		{
			name: "validation error (invalid offset)",
			filter: entities.TransactionFilter{
				Type:   entities.Win,
				UserID: userID,
			},
			orderBy: "amount",
			limit:   5,
			offset:  -2,
			mockResp: &txProto.GetTransactionByFiltersResponse{
				Transaction: []*txProto.Transaction{
					{Id: "invalid-uuid", Amount: 100},
				},
			},
			mockErr:       status.Error(codes.InvalidArgument, "invalid offset"),
			expectedErr:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedProtoType := transactionTypeEntityToProto[tt.filter.Type]

			mockCli.
				On("GetTransactionByFilters",
					mock.Anything,
					mock.MatchedBy(func(req *txProto.GetTransactionByFiltersRequest) bool {
						return req.OrderBy == tt.orderBy &&
							req.Limit == tt.limit &&
							req.Offset == tt.offset &&
							req.Filters != nil &&
							req.Filters.UserId == tt.filter.UserID &&
							req.Filters.Type == expectedProtoType
					}),
				).
				Return(tt.mockResp, tt.mockErr).
				Once()

			result, count, err := client.GetTransactions(
				context.Background(),
				tt.filter,
				tt.orderBy,
				tt.limit,
				tt.offset,
			)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockCli.AssertExpectations(t)
		})
	}
}
