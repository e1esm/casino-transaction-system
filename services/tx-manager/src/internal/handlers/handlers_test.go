package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers/mocks"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetTransactionByID(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name               string
		req                *proto.GetTransactionByIDRequest
		expectedResp       *proto.GetTransactionByIDResponse
		expectedStatusCode codes.Code
	}{
		{
			name: "invalid requested id",
			req: &proto.GetTransactionByIDRequest{
				Id: "some-invalid-id",
			},
			expectedResp:       nil,
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "not found error",
			req: &proto.GetTransactionByIDRequest{
				Id: uuid.NewString(),
			},
			expectedResp:       nil,
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "found transaction by id",
			req: &proto.GetTransactionByIDRequest{
				Id: id.String(),
			},
			expectedResp: &proto.GetTransactionByIDResponse{
				Transaction: &proto.Transaction{
					Id: id.String(),
				},
			},
			expectedStatusCode: codes.OK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cliMock := mocks.NewMockTransactionService(t)
			h := New(cliMock)

			switch {
			case test.expectedStatusCode == codes.InvalidArgument:
			case test.expectedStatusCode == codes.NotFound:
				cliMock.On("GetByID", context.Background(), uuid.MustParse(test.req.GetId())).Return(nil, svcerr.ErrNotFound)
			case test.expectedStatusCode == codes.OK:
				returnedTx := &models.Transaction{
					ID: id,
				}
				cliMock.On("GetByID", context.Background(), uuid.MustParse(test.req.GetId())).Return(returnedTx, nil)
			}

			resp, err := h.GetTransactionByID(context.Background(), test.req)
			if err != nil {
				assert.Nil(t, nil)
			} else {
				assert.Equal(t, test.expectedResp.Transaction.Id, resp.Transaction.Id, test.name)
			}

			assert.Equal(t, test.expectedStatusCode, status.Code(err))

			cliMock.AssertExpectations(t)
		})
	}
}

func TestHandler_GetTransactionByFilters(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	invalidArgFunc := func(err error) bool {
		if err == nil {
			return false
		}
		st, ok := status.FromError(err)
		if !ok {
			return false
		}
		return st.Code() == codes.InvalidArgument
	}

	tx1 := models.Transaction{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Type:            models.Bet,
		Amount:          100,
		TransactionTime: now,
	}
	tx2 := models.Transaction{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Type:            models.Win,
		Amount:          200,
		TransactionTime: now,
	}

	tests := []struct {
		name          string
		req           *proto.GetTransactionByFiltersRequest
		mockSetup     func(txSvc *mocks.MockTransactionService)
		wantErr       bool
		wantTotal     int64
		wantTxCount   int
		expectedErrFn func(error) bool
	}{
		{
			name: "valid request returns transactions",
			req: &proto.GetTransactionByFiltersRequest{
				Filters: &proto.Filters{
					Type: proto.TransactionType_Bet,
				},
				OrderBy: "amount",
				Limit:   10,
				Offset:  0,
			},
			mockSetup: func(txSvc *mocks.MockTransactionService) {
				txSvc.On("GetAll", mock.Anything, mock.Anything, "amount", int64(10), int64(0)).
					Return([]models.Transaction{tx1, tx2}, int64(2), nil)
			},
			wantErr:     false,
			wantTotal:   2,
			wantTxCount: 2,
		},
		{
			name: "invalid filters returns CastInvalidRequest",
			req: &proto.GetTransactionByFiltersRequest{
				Filters: &proto.Filters{
					UserId: "invalid-uuid",
				},
				OrderBy: "amount",
				Limit:   10,
				Offset:  0,
			},
			mockSetup:     func(txSvc *mocks.MockTransactionService) {},
			wantErr:       true,
			expectedErrFn: invalidArgFunc,
		},
		{
			name: "invalid limit/offset returns CastInvalidRequest",
			req: &proto.GetTransactionByFiltersRequest{
				Filters: &proto.Filters{},
				Limit:   0,
				Offset:  -1,
			},
			mockSetup:     func(txSvc *mocks.MockTransactionService) {},
			wantErr:       true,
			expectedErrFn: invalidArgFunc,
		},
		{
			name: "service returns error -> mapped error",
			req: &proto.GetTransactionByFiltersRequest{
				Filters: &proto.Filters{},
				Limit:   10,
				Offset:  0,
			},
			mockSetup: func(txSvc *mocks.MockTransactionService) {
				txSvc.On("GetAll", mock.Anything, mock.Anything, "", int64(10), int64(0)).
					Return(nil, int64(0), errors.New("internal service error"))
			},
			wantErr: true,
			expectedErrFn: func(err error) bool {
				st, ok := status.FromError(err)
				return ok && st.Code() == codes.Internal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txSvcMock := &mocks.MockTransactionService{}
			tt.mockSetup(txSvcMock)

			h := &Handler{
				txSvc: txSvcMock,
			}

			resp, err := h.GetTransactionByFilters(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrFn != nil {
					assert.True(t, tt.expectedErrFn(err))
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantTotal, resp.Total)
				assert.Len(t, resp.Transaction, tt.wantTxCount)
			}

			txSvcMock.AssertExpectations(t)
		})
	}
}
