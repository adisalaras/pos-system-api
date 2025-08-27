package main

import (
	"database/sql"
	"time"
)

type CategoryRepository interface {
	GetAll() ([]*Category, error) // ambil semua kategori
	GetByID(id int) (*Category, error) // ambil kategori berdasarkan ID
	Create(category *Category) error  // buat kategori baru
	Update(id int, category *Category) error // update kategori berdasarkan ID
	Delete(id int) error // hapus kategori berdasarkan ID
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll() ([]*Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description, 
			&category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *categoryRepository) GetByID(id int) (*Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`
	category := &Category{}
	err := r.db.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.Description,
		&category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) Create(category *Category) error {
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) 
			  RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, category.Name, category.Description).Scan(
		&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *categoryRepository) Update(id int, category *Category) error {
	query := `UPDATE categories SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(query, category.Name, category.Description, id).Scan(&category.UpdatedAt)
}

func (r *categoryRepository) Delete(id int) error {
	query := `DELETE FROM categories WHERE id = $1`
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