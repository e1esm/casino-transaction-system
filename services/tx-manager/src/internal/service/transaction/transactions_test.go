package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/service/transaction/mocks"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceCreate(t *testing.T) {
	cliMock := mocks.NewMockRepository(t)
	svc := New(cliMock)

	tests := []struct {
		name         string
		transactions []models.Transaction
		expectedErr  error
	}{
		{
			name:         "empty transactions",
			transactions: []models.Transaction{},
			expectedErr:  nil,
		},
		{
			name: "single transaction",
			transactions: []models.Transaction{
				{
					UserID:          uuid.New(),
					Amount:          100,
					Type:            models.Bet,
					TransactionTime: time.Now(),
				},
			},
			expectedErr: nil,
		},
		{
			name:         "nil transaction slice",
			transactions: nil,
			expectedErr:  nil,
		},
		{
			name: "multiple transactions",
			transactions: []models.Transaction{
				{
					UserID:          uuid.New(),
					Amount:          100,
					Type:            models.Bet,
					TransactionTime: time.Now(),
				},
				{
					UserID:          uuid.New(),
					Amount:          20,
					Type:            models.Bet,
					TransactionTime: time.Now(),
				},
				{
					UserID:          uuid.New(),
					Amount:          35,
					Type:            models.Bet,
					TransactionTime: time.Now(),
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		if len(tt.transactions) != 0 {
			cliMock.On("Add", mock.Anything, mock.Anything).Return(tt.expectedErr)
		}

		err := svc.Create(context.Background(), tt.transactions...)

		assert.Equal(t, tt.expectedErr, err, tt.name)
		cliMock.AssertExpectations(t)
	}
}

func TestServiceGetByID(t *testing.T) {
	cliMock := mocks.NewMockRepository(t)
	svc := New(cliMock)
	id := uuid.New()

	tests := []struct {
		name        string
		requestedID uuid.UUID
		expectedTx  *models.Transaction
		expectedErr error
	}{
		{
			name:        "transaction not found",
			requestedID: uuid.New(),
			expectedTx:  nil,
			expectedErr: svcerr.ErrNotFound,
		},
		{
			name:        "transaction was found",
			requestedID: id,
			expectedTx: &models.Transaction{
				ID:              id,
				UserID:          uuid.Nil,
				Amount:          100,
				Type:            models.Bet,
				TransactionTime: time.Now(),
			},
			expectedErr: nil,
		},
		{
			name:        "some internal error",
			requestedID: id,
			expectedTx:  nil,
			expectedErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		switch {
		case errors.Is(tt.expectedErr, svcerr.ErrNotFound):
			cliMock.On("GetByID", mock.Anything, tt.requestedID).Return(tt.expectedTx, nil)
		case tt.expectedErr == nil:
			cliMock.On("GetByID", mock.Anything, tt.requestedID).Return(tt.expectedTx, nil)
		default:
			cliMock.On("GetByID", mock.Anything, tt.requestedID).Return(nil, tt.expectedErr)

		}

		resp, err := svc.GetByID(context.Background(), tt.requestedID)
		assert.Equal(t, tt.expectedTx, resp, tt.name)
		assert.ErrorIs(t, err, tt.expectedErr, tt.name)
		cliMock.AssertExpectations(t)
		cliMock.ExpectedCalls = nil
	}
}

func TestServiceGetAll(t *testing.T) {
	cliMock := mocks.NewMockRepository(t)
	svc := New(cliMock)

	tests := []struct {
		name          string
		filters       models.TransactionFilter
		orderBy       string
		offset, limit int64
		expectedTotal int64
		expectedTxs   []models.Transaction
		expectedErr   error
	}{
		{
			name:          "invalid order by",
			filters:       models.TransactionFilter{},
			orderBy:       "date desc",
			offset:        0,
			limit:         1,
			expectedTxs:   nil,
			expectedTotal: -1,
			expectedErr:   svcerr.ErrBadField,
		},
		{
			name:        "limit 0",
			filters:     models.TransactionFilter{},
			orderBy:     "",
			offset:      0,
			limit:       0,
			expectedTxs: nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		cliMock.On("GetAll", mock.Anything, tt.filters, tt.orderBy, tt.limit, tt.offset).Return(tt.expectedTxs, tt.expectedErr)

		resp, total, err := svc.GetAll(context.Background(), tt.filters, tt.orderBy, tt.limit, tt.offset)

		assert.Equal(t, tt.expectedTotal, total, tt.name)
		assert.Equal(t, tt.expectedTxs, resp, tt.name)
		assert.ErrorIs(t, err, tt.expectedErr, tt.name)

		cliMock.AssertExpectations(t)
		cliMock.ExpectedCalls = nil
	}
}
