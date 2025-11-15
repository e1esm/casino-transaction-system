package transaction

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:14-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
	)

	if err != nil {
		log.Fatalf("Could not start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("Error terminating container: %v", err)
		}
	}()

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Could not get connection string: %v", err)
	}

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Could not open database connection: %v", err)
	}
	testDB = db
	defer db.Close()

	_, err = testDB.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto;")
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open sql connection for migrations: %v", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("failed to set postgres dialect")
	}

	if err := goose.Up(sqlDB, "../../../../../../migrations/tx_manager"); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	code := m.Run()

	os.Exit(code)
}

func TestRepositoryAddIntegrations(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	repo := NewWithPool(testDB)

	_, _ = testDB.Exec(ctx, "DELETE FROM transactions")

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

	t.Run("add batch of transactions", func(t *testing.T) {
		err := repo.Add(ctx, tx1, tx2)
		assert.Nil(t, err)

		resp, err := repo.GetAll(ctx, models.TransactionFilter{}, "", 10, 0)
		assert.Nil(t, err)
		assert.Len(t, resp, 2)
	})

	t.Run("add duplicate transactions should not insert again", func(t *testing.T) {
		err := repo.Add(ctx, tx1)
		if err != nil {
			t.Fatalf("failed to add transaction: %v", err)
		}

		resp, err := repo.GetAll(ctx, models.TransactionFilter{}, "", 10, 0)
		assert.Nil(t, err)
		assert.Len(t, resp, 2)
	})
}

func TestRepositoryGetAllIntegration(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	repo := NewWithPool(testDB)

	_, _ = testDB.Exec(ctx, "DELETE FROM transactions")

	tx1 := models.Transaction{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Type:            models.Bet,
		Amount:          100,
		TransactionTime: now,
	}
	tx2 := models.Transaction{
		ID:              uuid.New(),
		UserID:          tx1.UserID,
		Type:            models.Win,
		Amount:          200,
		TransactionTime: now,
	}
	tx3 := models.Transaction{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Type:            models.Bet,
		Amount:          300,
		TransactionTime: now,
	}

	err := repo.Add(ctx, tx1, tx2, tx3)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		filter        models.TransactionFilter
		expectedCount int
		orderBy       string
		expectedErr   bool
	}{
		{"all transactions", models.TransactionFilter{}, 3, "id desc", true},
		{"filter by user", models.TransactionFilter{UserID: &tx1.UserID}, 2, "id", true},
		{"filter by type", models.TransactionFilter{Type: &tx3.Type}, 2, "user_id desc", false},
		{"filter by user and type", models.TransactionFilter{UserID: &tx1.UserID, Type: &tx1.Type}, 1, "amount", true},
		{"non-existing filter", models.TransactionFilter{UserID: &uuid.UUID{}}, 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetAll(ctx, tt.filter, tt.orderBy, 100, 0)
			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)
			assert.Len(t, results, tt.expectedCount)
		})
	}
}

func TestRepositoryGetByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	repo := NewWithPool(testDB)
	n := int64(10)

	_, err := testDB.Exec(ctx, "DELETE FROM transactions")
	assert.NoError(t, err)

	for range n {
		err = repo.Add(ctx, models.Transaction{
			ID:              uuid.New(),
			UserID:          uuid.New(),
			Type:            models.Bet,
			Amount:          rand.Intn(100),
			TransactionTime: now,
		})
		assert.NoError(t, err)
	}

	resp, err := repo.GetAll(ctx, models.TransactionFilter{}, "", n, 0)
	assert.NoError(t, err)
	assert.Len(t, resp, int(n))

	type tCase struct {
		id        uuid.UUID
		toBeFound bool
		dropConn  bool
	}

	tests := []tCase{
		{id: uuid.New(), toBeFound: false, dropConn: true},
	}

	for _, tr := range resp {
		tests = append(tests, tCase{
			id:        tr.ID,
			toBeFound: true,
		})
	}

	for range n {
		tests = append(tests, tCase{
			id:        uuid.New(),
			toBeFound: false,
		})
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			if tt.dropConn {
				repo.db, _ = pgxpool.New(ctx, "")
				defer func() {
					repo.db = testDB
				}()
			}

			tr, err := repo.GetByID(ctx, tt.id)
			if tt.toBeFound {
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, tt.id)

				return
			}

			if tt.dropConn {
				assert.Error(t, err)
				assert.Nil(t, tr)
				return
			}

			assert.NoError(t, err)
			assert.Nil(t, tr)
		})
	}
}
