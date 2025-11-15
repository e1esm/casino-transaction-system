package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	mocks "github.com/e1esm/casino-transaction-system/api-gateway/src/internal/handlers/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		name             string
		status           int
		errMsg           string
		expectedStatus   int
		expectedResponse map[string]string
	}{
		{
			name:             "simple error",
			status:           http.StatusBadRequest,
			errMsg:           "invalid payload",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: map[string]string{"error": "invalid payload"},
		},
		{
			name:             "not found",
			status:           http.StatusNotFound,
			errMsg:           "resource not found",
			expectedStatus:   http.StatusNotFound,
			expectedResponse: map[string]string{"error": "resource not found"},
		},
		{
			name:             "empty error message",
			status:           http.StatusInternalServerError,
			errMsg:           "",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: map[string]string{"error": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			writeJSONError(rr, tt.status, tt.errMsg)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %s", ct)
			}

			var got map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			if got["error"] != tt.expectedResponse["error"] {
				t.Fatalf("expected error message %q, got %q",
					tt.expectedResponse["error"], got["error"])
			}
		})
	}
}

func TestHandler_GetTransactions(t *testing.T) {
	cliMock := mocks.NewMockClient(t)
	h := New(cliMock)

	tests := []struct {
		name           string
		query          string
		queryFilters   entities.TransactionFilter
		mockReturn     []entities.Transaction
		mockTotal      int
		mockErr        error
		expectedStatus int
	}{
		{
			name:  "success with filters",
			query: "?filters=%7B%22type%22%3A%20%22win%22%7D",
			queryFilters: entities.TransactionFilter{
				Type: "win",
			},
			mockReturn:     []entities.Transaction{{ID: uuid.New(), Type: entities.Win, Amount: 33}},
			mockTotal:      1,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success",
			query:          "",
			mockReturn:     []entities.Transaction{{ID: uuid.New(), Amount: 100}, {ID: uuid.New(), Amount: 200}},
			mockTotal:      2,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bad filters",
			query:          "?filters=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal server error",
			query:          "",
			mockErr:        errors.New("internal"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErr == nil && tt.expectedStatus == http.StatusOK {
				cliMock.On("GetTransactions", mock.Anything, tt.queryFilters, "", int64(10), int64(0)).
					Return(tt.mockReturn, tt.mockTotal, nil).Once()
			} else if tt.mockErr != nil {
				cliMock.On("GetTransactions", mock.Anything, tt.queryFilters, "", int64(10), int64(0)).
					Return(nil, 0, tt.mockErr).Once()
			}

			req := httptest.NewRequest(http.MethodGet, "/transactions"+tt.query, nil)
			w := httptest.NewRecorder()

			h.GetTransactions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
			cliMock.AssertExpectations(t)
			cliMock.ExpectedCalls = nil
		})
	}
}

func TestHandler_GetTransactionByID(t *testing.T) {
	cliMock := mocks.NewMockClient(t)
	h := New(cliMock)

	validID := uuid.New()
	tests := []struct {
		name           string
		id             string
		mockReturn     entities.Transaction
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			id:             validID.String(),
			mockReturn:     entities.Transaction{ID: validID, Amount: 500},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid id",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal error",
			id:             validID.String(),
			mockErr:        errors.New("internal"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErr == nil && tt.expectedStatus == http.StatusOK {
				cliMock.On("GetTransactionByID", mock.Anything, validID).Return(tt.mockReturn, nil).Once()
			} else if tt.mockErr != nil {
				cliMock.On("GetTransactionByID", mock.Anything, validID).Return(entities.Transaction{}, tt.mockErr).Once()
			}

			req := httptest.NewRequest(http.MethodGet, "/transactions/", nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			h.GetTransactionByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
			cliMock.AssertExpectations(t)
			cliMock.ExpectedCalls = nil
		})
	}
}
