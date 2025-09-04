package storer

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PySQLStorer struct {
	db *sqlx.DB
}

func NewPySQLStorer(db *sqlx.DB) *PySQLStorer {
	return &PySQLStorer{db: db}
}

func (ps *PySQLStorer) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = &now

	var id int64
	err := ps.db.QueryRowxContext(
		ctx,
		`INSERT INTO products (
            name, image, category, description, rating, num_reviews, price, count_in_stock, created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING id`,
		p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews,
		p.Price, p.CountInStock, p.CreatedAt, p.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	p.ID = id
	return p, nil
}

func (ps *PySQLStorer) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := ps.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product with id %d: %w", id, err)
	}
	return &p, nil
}

func (ps *PySQLStorer) ListProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	err := ps.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

func (ps *PySQLStorer) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	rows, err := ps.db.NamedQueryContext(
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
		WHERE id = :id
		RETURNING *`,
		p,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product with id %d: %w", p.ID, err)
	}
	defer rows.Close()

	if rows.Next() {
		var updated Product
		if err := rows.StructScan(&updated); err != nil {
			return nil, fmt.Errorf("failed to scan updated product: %w", err)
		}
		return &updated, nil
	}

	return nil, fmt.Errorf("no product found with id %d", p.ID)
}

func (ps *PySQLStorer) DeleteProduct(ctx context.Context, id int64) error {
	_, err := ps.db.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product with id %d: %w", id, err)
	}
	return nil
}

func (ps *PySQLStorer) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = &now

	err := ps.execTx(ctx, func(tx *sqlx.Tx) error {
		// insert into orders
		createdOrder, err := createOrder(ctx, tx, o)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, oi := range o.Items {
			oi.OrderID = createdOrder.ID
			// insert into order_items
			_, err := createOrderItem(ctx, tx, &oi)
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return o, nil
}

func createOrder(ctx context.Context, tx *sqlx.Tx, o *Order) (*Order, error) {
	err := tx.QueryRowxContext(
		ctx,
		`INSERT INTO orders (
			payment_method, tax_price, shipping_price, total_price, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.CreatedAt, o.UpdatedAt,
	).Scan(&o.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}
	return o, nil
}

func createOrderItem(ctx context.Context, tx *sqlx.Tx, oi *OrderItem) (*OrderItem, error) {
	err := tx.QueryRowxContext(
		ctx,
		`INSERT INTO order_items (
			name, quantity, image, price, product_id, order_id
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		oi.Name, oi.Quantity, oi.Image, oi.Price, oi.ProductID, oi.OrderID,
	).Scan(&oi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order item: %w", err)
	}
	return oi, nil
}

func (ps *PySQLStorer) GetOrder(ctx context.Context, id int64) (*Order, error) {
	var o Order
	err := ps.db.GetContext(ctx, &o, "SELECT * FROM orders WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order with id %d: %w", id, err)
	}

	var items []OrderItem
	err = ps.db.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items for order id %d: %w", id, err)
	}
	o.Items = items

	return &o, nil
}

func (ps *PySQLStorer) ListOrders(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := ps.db.SelectContext(ctx, &orders, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	for i := range orders {
		var items []OrderItem
		err = ps.db.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=$1", orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order items for order id: %w", err)
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (ps *PySQLStorer) DeleteOrder(ctx context.Context, id int64) error {
	err := ps.execTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id=$1", id)
		if err != nil {
			return fmt.Errorf("failed to delete order items for order id %d: %w", id, err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id=$1", id)
		if err != nil {
			return fmt.Errorf("failed to delete order with id %d: %w", id, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete order with id %d: %w", id, err)
	}

	return nil
}

func (ps *PySQLStorer) execTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := ps.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return fmt.Errorf("erro in transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
