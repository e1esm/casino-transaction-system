package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/handlers/errors"
	"github.com/google/uuid"
)

type Client interface {
	GetTransactionByID(ctx context.Context, id uuid.UUID) (entities.Transaction, error)
	GetTransactions(ctx context.Context, filter entities.TransactionFilter, orderBy string, limit, offset int64) ([]entities.Transaction, int, error)
}

type Handler struct {
	cli Client
}

func New(cli Client) *Handler {
	return &Handler{
		cli: cli,
	}
}

// GetTransactions godoc
// @Summary Get a list of transactions
// @Description Returns transactions with optional filtering, pagination, and ordering
// @Tags transactions
// @Accept json
// @Produce json
// @Param limit query int false "Number of transactions to return" default(10)
// @Param offset query int false "Pagination offset" default(0)
// @Param orderBy query string false "Field to order by, e.g., amount desc"
// @Param filters query string false "JSON-encoded filters, e.g., {\"user_id\":\"uuid\",\"type\":\"bet\"}"
// @Success 200 {object} transactions "Transactions list and total count"
// @Failure 400 {object} string "Invalid request parameters"
// @Failure 500 {object} string "Internal server error"
// @Router /transactions [get]
func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	limit := strToIntWithDefault(r.URL.Query().Get("limit"), 10)
	offset := strToIntWithDefault(r.URL.Query().Get("offset"), 0)
	orderBy := r.URL.Query().Get("orderBy")

	filters, err := parseFiltersStruct(r.URL.Query().Get("filters"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid filters parameter")
		return
	}

	trResp, total, err := h.cli.GetTransactions(r.Context(), filters, orderBy, limit, offset)
	if err != nil {
		code, errMsg := errors.ParseSvcErrToResp(err)
		if code == http.StatusInternalServerError {
			log.Println(err.Error())
		}

		writeJSONError(w, code, errMsg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(convertTransactionEntitiesToResponse(trResp, total)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// GetTransactionByID godoc
// @Summary Get a single transaction by ID
// @Description Returns a transaction by its UUID
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} transaction "Transaction object"
// @Failure 400 {object} string "Invalid or missing ID"
// @Failure 404 {object} string "Transaction not found"
// @Failure 500 {object} string "Internal server error"
// @Router /transactions/{id} [get]
func (h *Handler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing id parameter")
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id parameter")
		return
	}

	resp, err := h.cli.GetTransactionByID(r.Context(), parsedID)
	if err != nil {
		code, errMsg := errors.ParseSvcErrToResp(err)
		if code == http.StatusInternalServerError {
			log.Println(err.Error())
		}

		writeJSONError(w, code, errMsg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(convertTransactionEntityToResponse(resp)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
