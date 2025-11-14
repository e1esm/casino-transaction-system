package transaction

import (
	"context"
	"fmt"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"
	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	GetAll(ctx context.Context, filters models.TransactionFilter, orderBy string, limit, offset int64) ([]models.Transaction, error)
	Add(ctx context.Context, transactions ...models.Transaction) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	resp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("transaction with such id was not found: %w", svcerr.ErrNotFound)
	}

	return resp, nil
}

func (s *Service) GetAll(ctx context.Context, filters models.TransactionFilter, orderBy string, limit, offset int64) ([]models.Transaction, int64, error) {
	resp, err := s.repo.GetAll(ctx, filters, orderBy, limit, offset)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to get transactions by filters: %w", err)
	}

	return resp, int64(len(resp)), nil
}

func (s *Service) Create(ctx context.Context, transactions ...models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	return s.repo.Add(ctx, transactions...)
}
