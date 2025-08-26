package storer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PySQLStorer struct {
	db *sqlx.DB
}

func NewPySQLStorer(db *sqlx.DB) *PySQLStorer {
	return &PySQLStorer{db: db}
}

func (ms *PySQLStorer) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	res, err := ms.db.NamedExecContext(
		ctx,
		`INSERT INTO products (
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
        :name, 
        :image, 
        :category, 
        :description, 
        :rating, 
        :num_reviews, 
        :price, 
        :count_in_stock, 
        :created_at, 
        :updated_at
    )`,
		p,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	p.ID = id

	return p, nil
}

func (ms *PySQLStorer) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := ms.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id=?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product with id %d: %w", id, err)
	}
	return &p, nil
}

func (ms *PySQLStorer) ListProduct(ctx context.Context, p *Product) ([]*Product, error) {
	var products []*Product
	err := ms.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

func (ms *PySQLStorer) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	_, err := ms.db.NamedExecContext(
		ctx,
		`UPDATE products SET 
		name = :name, 
		image = :image, 
		category = :category, 
		description = :description, 
		rating = :rating, 
		num_reviews = :num_reviews, 
		price = :price, 
		count_in_stock = :count_in_stock, 
		updated_at = :updated_at 
	WHERE id = :id`,
		p,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product with id %d: %w", p.ID, err)
	}

	return p, nil
}

func (ms *PySQLStorer) DeleteProduct(ctx context.Context, id int64) error {
	_, err := ms.db.ExecContext(ctx, "DELETE FROM products WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete product with id %d: %w", id, err)
	}
	return nil
}
