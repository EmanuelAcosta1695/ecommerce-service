package storer

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	st := NewPySQLStorer(db)

	p := &Product{
		Name:         "Test Product",
		Image:        "test_image.jpg",
		Category:     "Test Category",
		Description:  "Test Description",
		Rating:       5,
		NumReviews:   10,
		Price:        99.99,
		CountInStock: 100,
		CreatedAt:    time.Now(),
		UpdatedAt:    nil,
	}

	mock.ExpectExec(`INSERT INTO products (
		name,
		image,
		category,
		description,
		rating,
		num_reviews,
		price,
		count_in_stock,
		created_at,
		updated_at
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)`).WillReturnResult(sqlmock.NewResult(1, 1))

	cp, err := st.CreateProduct(context.Background(), p)
	require.NoError(t, err)
	require.Equal(t, int64(1), cp.ID)
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
