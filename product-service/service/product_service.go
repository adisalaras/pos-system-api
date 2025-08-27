package main

import (
	"fmt"
)

type ProductService interface {
	GetAllProducts(filters ProductFilters) ([]*Product, error)
	GetProductByID(id int) (*Product, error)
	GetProductBySKU(sku string) (*Product, error)
	CreateProduct(req CreateProductRequest) (*Product, error)
	UpdateProduct(id int, req UpdateProductRequest) (*Product, error)
	UpdateStock(id int, req StockUpdateRequest) error
	DeleteProduct(id int) error
	GetLowStockProducts() ([]*Product, error)
	CheckProductStock(id int, quantity int) (bool, error)
}

type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetAllProducts(filters ProductFilters) ([]*Product, error) {
	return s.repo.GetAll(filters)
}

func (s *productService) GetProductByID(id int) (*Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) GetProductBySKU(sku string) (*Product, error) {
	return s.repo.GetBySKU(sku)
}

func (s *productService) CreateProduct(req CreateProductRequest) (*Product, error) {
	// Validasi input
	if req.Name == "" {
		return nil, fmt.Errorf("product name wajib diisi")
	}
	if req.SKU == "" {
		return nil, fmt.Errorf("product SKU wajib diisi")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("price tidak boleh negatif")
	}

	product := &Product{
		Name:          req.Name,
		SKU:           req.SKU,
		CategoryID:    req.CategoryID,
		Price:         req.Price,
		Cost:          req.Cost,
		StockQuantity: req.StockQuantity,
		MinStock:      req.MinStock,
		Description:   req.Description,
		ImageURL:      req.ImageURL,
		IsActive:      true,
	}

	err := s.repo.Create(product)
	if err != nil {
		return nil, fmt.Errorf("produk gagal dibuat: %w", err)
	}

	return product, nil
}

func (s *productService) UpdateProduct(id int, req UpdateProductRequest) (*Product, error) {
	// Ambil produk yang ada menggunakan id
	existingProduct, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("produk tidak ditemukan: %w", err)
	}

	// Update data produk
	if req.Name != nil {
		existingProduct.Name = *req.Name
	}
	if req.CategoryID != nil {
		existingProduct.CategoryID = req.CategoryID
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, fmt.Errorf("price tidak boleh negatif")
		}
		existingProduct.Price = *req.Price
	}
	if req.Cost != nil {
		if *req.Cost < 0 {
			return nil, fmt.Errorf("cost tidak boleh negatif")
		}
		existingProduct.Cost = *req.Cost
	}
	if req.StockQuantity != nil {
		if *req.StockQuantity < 0 {
			return nil, fmt.Errorf("stock quantity tidak boleh negatif")
		}
		existingProduct.StockQuantity = *req.StockQuantity
	}
	if req.MinStock != nil {
		if *req.MinStock < 0 {
			return nil, fmt.Errorf("minimum stock tidak boleh negatif")
		}
		existingProduct.MinStock = *req.MinStock
	}
	if req.Description != nil {
		existingProduct.Description = req.Description
	}
	if req.ImageURL != nil {
		existingProduct.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		existingProduct.IsActive = *req.IsActive
	}

	err = s.repo.Update(id, existingProduct)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui produk: %w", err)
	}

	return existingProduct, nil
}

func (s *productService) UpdateStock(id int, req StockUpdateRequest) error {
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity harus lebih dari nol")
	}

	if req.Type != "add" && req.Type != "subtract" {
		return fmt.Errorf("tipe stock tidak valid: %s", req.Type)
	}

	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("produk tidak ditemukan: %w", err)
	}

	if req.Type == "subtract" {
		hasStock, err := s.repo.CheckStock(id, req.Quantity)
		if err != nil {
			return fmt.Errorf("gagal memeriksa stock: %w", err)
		}
		if !hasStock {
			return fmt.Errorf("stock tidak mencukupi")
		}
	}

	err = s.repo.UpdateStock(id, req.Quantity, req.Type)
	if err != nil {
		return fmt.Errorf("gagal memperbarui stock: %w", err)
	}

	return nil
}

func (s *productService) DeleteProduct(id int) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus produk: %w", err)
	}
	return nil
}

func (s *productService) GetLowStockProducts() ([]*Product, error) {
	return s.repo.GetLowStock()
}

func (s *productService) CheckProductStock(id int, quantity int) (bool, error) {
	return s.repo.CheckStock(id, quantity)
}

