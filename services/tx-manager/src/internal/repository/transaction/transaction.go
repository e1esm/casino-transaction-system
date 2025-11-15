package transaction

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/config"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var orderByFields = map[string]struct{}{
	"user_id":          {},
	"amount":           {},
	"transaction_type": {},
	"timestamp":        {},
}

type Repository struct {
	db *pgxpool.Pool
}

func NewWithPool(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func New(cfg config.DatabaseConfig) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
			cfg.SSLMode),
	)

	if err != nil {
		return nil, err
	}

	return NewWithPool(pool), nil
}

func (r *Repository) Add(ctx context.Context, transactions ...models.Transaction) error {
	query := `
        INSERT INTO transactions (user_id, transaction_type, amount, transaction_time, t_hash)
        VALUES ($1, $2, $3, $4, $5) ON CONFLICT (t_hash) DO NOTHING
    `

	batch := &pgx.Batch{}
	for _, t := range transactions {
		batch.Queue(query, t.UserID, t.Type, t.Amount, t.TransactionTime, t.Hash())
	}

	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	_, err := batchResults.Exec()

	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	resp := models.Transaction{}

	query := `
		SELECT id, user_id, transaction_type, amount, transaction_time FROM transactions
		where id = $1
    `

	err := r.db.QueryRow(ctx, query, id).Scan(&resp.ID, &resp.UserID, &resp.Type, &resp.Amount, &resp.TransactionTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &resp, nil
}

func (r *Repository) GetAll(ctx context.Context, filters models.TransactionFilter, orderBy string, limit, offset int64) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, transaction_type, amount, transaction_time FROM transactions
    `

	cond, args := filters.String()
	if len(cond) > 0 {
		query += " WHERE " + cond
	}

	if len(orderBy) > 0 {
		if !validateOrderBy(orderBy) {
			return nil, fmt.Errorf("%w: invalid orderBy: %s", svcerr.ErrBadField, orderBy)
		}

		query += fmt.Sprintf(" ORDER BY %s", orderBy)
	}

	args = append(args, limit, offset)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	var resp []models.Transaction
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var t models.Transaction

		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.TransactionTime); err != nil {
			return nil, err
		}

		resp = append(resp, t)
	}

	return resp, nil
}

func (r *Repository) Close() {
	r.db.Close()
}

func validateOrderBy(orderBy string) bool {
	parts := strings.Split(orderBy, " ")
	if len(parts) != 2 {
		return false
	}

	_, ok := orderByFields[parts[0]]
	return ok
}
