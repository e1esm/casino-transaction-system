package handlers

import (
	"context"
	"errors"
	"log"

	hErr "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers/errors"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/handlers/validators"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	proto "github.com/e1esm/casino-transaction-system/tx-manager/src/internal/proto/tx-manager"

	"github.com/google/uuid"
)

type TransactionService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	GetAll(ctx context.Context, filters models.TransactionFilter, orderBy string, limit, offset int64) ([]models.Transaction, int64, error)
}

type Handler struct {
	proto.UnimplementedTransactionManagerServer

	txSvc TransactionService
}

func New(txSvc TransactionService) *Handler {
	return &Handler{txSvc: txSvc}
}

func (h *Handler) GetTransactionByID(ctx context.Context, req *proto.GetTransactionByIDRequest) (*proto.GetTransactionByIDResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, hErr.CastInvalidRequest(err)
	}

	resp, err := h.txSvc.GetByID(ctx, id)
	if err != nil {
		prErr, isInternal := hErr.ParseSvcErrToProto(err)
		if isInternal {
			log.Println(prErr.Error())
		}
		return nil, prErr
	}

	return &proto.GetTransactionByIDResponse{
		Transaction: convertTransactionModelToProto(*resp),
	}, nil
}

func (h *Handler) GetTransactionByFilters(ctx context.Context, req *proto.GetTransactionByFiltersRequest) (*proto.GetTransactionByFiltersResponse, error) {
	parsedFilters, err := convertProtoFiltersToModel(req.Filters)
	if err != nil {
		return nil, hErr.CastInvalidRequest(err)
	}

	if !validators.ValidateGreaterOrEqualTo(1, req.Limit) || !validators.ValidateGreaterOrEqualTo(0, req.Offset) {
		return nil, hErr.CastInvalidRequest(errors.New("invalid offset or limit"))
	}

	resp, n, err := h.txSvc.GetAll(ctx, parsedFilters, req.OrderBy, req.Limit, req.Offset)
	if err != nil {
		prErr, isInternal := hErr.ParseSvcErrToProto(err)
		if isInternal {
			log.Println(prErr.Error())
		}

		return nil, prErr
	}

	return &proto.GetTransactionByFiltersResponse{
		Transaction: convertTransactionsModelToProto(resp),
		Total:       n,
	}, nil
}
