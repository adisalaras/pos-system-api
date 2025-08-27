package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type ProductRepository interface {
	GetAll(filters ProductFilters) ([]*Product, error)
	GetByID(id int) (*Product, error)
	GetBySKU(sku string) (*Product, error)
	Create(product *Product) error
	Update(id int, product *Product) error
	UpdateStock(id int, quantity int, operation string) error
	Delete(id int) error
	GetLowStock() ([]*Product, error)
	CheckStock(id int, quantity int) (bool, error)
}

type ProductFilters struct {
	CategoryID *int
	Search     string
	IsActive   *bool
	LowStock   bool
	Limit      int
	Offset     int
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(filters ProductFilters) ([]*Product, error) {
	query := `SELECT p.id, p.name, p.sku, p.category_id, p.price, p.cost, p.stock_quantity, p.min_stock, p.description, p.image_url, p.is_active,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.created_at, c.updated_at
			  FROM products p 
			  LEFT JOIN categories c ON p.category_id = c.id`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *filters.CategoryID)
		argIndex++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.LowStock {
		conditions = append(conditions, "p.stock_quantity <= p.min_stock")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY p.name"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		var categoryID, catID sql.NullInt64
		var catName, catDesc sql.NullString
		var catCreatedAt, catUpdatedAt sql.NullTime

		err := rows.Scan(
			&product.ID, &product.Name, &product.SKU, &categoryID,
			&product.Price, &product.Cost, &product.StockQuantity, &product.MinStock,
			&product.Description, &product.ImageURL, &product.IsActive,
			&product.CreatedAt, &product.UpdatedAt,
			&catID, &catName, &catDesc, &catCreatedAt, &catUpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if categoryID.Valid {
			id := int(categoryID.Int64)
			product.CategoryID = &id
		}

		if catID.Valid {
			product.Category = &Category{
				ID:          int(catID.Int64),
				Name:        catName.String,
				Description: &catDesc.String,
				CreatedAt:   catCreatedAt.Time,
				UpdatedAt:   catUpdatedAt.Time,
			}
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *productRepository) GetByID(id int) (*Product, error) {
	query := `SELECT p.id, p.name, p.sku, p.category_id, p.price, p.cost, 
			  p.stock_quantity, p.min_stock, p.description, p.image_url, p.is_active,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.created_at, c.updated_at
			  FROM products p 
			  LEFT JOIN categories c ON p.category_id = c.id 
			  WHERE p.id = $1`

	product := &Product{}
	var categoryID, catID sql.NullInt64
	var catName, catDesc sql.NullString
	var catCreatedAt, catUpdatedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.SKU, &categoryID,
		&product.Price, &product.Cost, &product.StockQuantity, &product.MinStock,
		&product.Description, &product.ImageURL, &product.IsActive,
		&product.CreatedAt, &product.UpdatedAt,
		&catID, &catName, &catDesc, &catCreatedAt, &catUpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if categoryID.Valid {
		id := int(categoryID.Int64)
		product.CategoryID = &id
	}

	if catID.Valid {
		product.Category = &Category{
			ID:          int(catID.Int64),
			Name:        catName.String,
			Description: &catDesc.String,
			CreatedAt:   catCreatedAt.Time,
			UpdatedAt:   catUpdatedAt.Time,
		}
	}

	return product, nil
}

func (r *productRepository) GetBySKU(sku string) (*Product, error) {
	query := `SELECT p.id, p.name, p.sku, p.category_id, p.price, p.cost, 
			  p.stock_quantity, p.min_stock, p.description, p.image_url, p.is_active,
			  p.created_at, p.updated_at
			  FROM products p WHERE p.sku = $1`

	product := &Product{}
	var categoryID sql.NullInt64

	err := r.db.QueryRow(query, sku).Scan(
		&product.ID, &product.Name, &product.SKU, &categoryID,
		&product.Price, &product.Cost, &product.StockQuantity, &product.MinStock,
		&product.Description, &product.ImageURL, &product.IsActive,
		&product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if categoryID.Valid {
		id := int(categoryID.Int64)
		product.CategoryID = &id
	}

	return product, nil
}

func (r *productRepository) Create(product *Product) error {
	query := `INSERT INTO products (name, sku, category_id, price, cost, stock_quantity, min_stock, description, image_url) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
			  RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, product.Name, product.SKU, product.CategoryID, product.Price,
		product.Cost, product.StockQuantity, product.MinStock, product.Description, product.ImageURL).
		Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *productRepository) Update(id int, product *Product) error {
	query := `UPDATE products SET name = $1, category_id = $2, price = $3, cost = $4, 
			  stock_quantity = $5, min_stock = $6, description = $7, image_url = $8, 
			  is_active = $9, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $10 RETURNING updated_at`
	return r.db.QueryRow(query, product.Name, product.CategoryID, product.Price, product.Cost,
		product.StockQuantity, product.MinStock, product.Description, product.ImageURL,
		product.IsActive, id).Scan(&product.UpdatedAt)
}

func (r *productRepository) UpdateStock(id int, quantity int, operation string) error {
	var query string
	switch operation {
	case "add":
		query = `UPDATE products SET stock_quantity = stock_quantity + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	case "subtract":
		query = `UPDATE products SET stock_quantity = stock_quantity - $1, updated_at = CURRENT_TIMESTAMP 
				 WHERE id = $2 AND stock_quantity >= $1`
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}

	result, err := r.db.Exec(query, quantity, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		if operation == "subtract" {
			return fmt.Errorf("insufficient stock")
		}
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepository) GetLowStock() ([]*Product, error) {
	query := `SELECT p.id, p.name, p.sku, p.category_id, p.price, p.cost, 
			  p.stock_quantity, p.min_stock, p.description, p.image_url, p.is_active,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.created_at, c.updated_at
			  FROM products p 
			  LEFT JOIN categories c ON p.category_id = c.id
			  WHERE p.stock_quantity <= p.min_stock AND p.is_active = true
			  ORDER BY (p.stock_quantity - p.min_stock), p.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		var categoryID, catID sql.NullInt64
		var catName, catDesc sql.NullString
		var catCreatedAt, catUpdatedAt sql.NullTime

		err := rows.Scan(
			&product.ID, &product.Name, &product.SKU, &categoryID,
			&product.Price, &product.Cost, &product.StockQuantity, &product.MinStock,
			&product.Description, &product.ImageURL, &product.IsActive,
			&product.CreatedAt, &product.UpdatedAt,
			&catID, &catName, &catDesc, &catCreatedAt, &catUpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if categoryID.Valid {
			id := int(categoryID.Int64)
			product.CategoryID = &id
		}

		if catID.Valid {
			product.Category = &Category{
				ID:          int(catID.Int64),
				Name:        catName.String,
				Description: &catDesc.String,
				CreatedAt:   catCreatedAt.Time,
				UpdatedAt:   catUpdatedAt.Time,
			}
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *productRepository) CheckStock(id int, quantity int) (bool, error) {
	query := `SELECT stock_quantity FROM products WHERE id = $1`
	var currentStock int
	err := r.db.QueryRow(query, id).Scan(&currentStock)
	if err != nil {
		return false, err
	}
	return currentStock >= quantity, nil
}