package main

import (
	"fmt"
)

type CategoryService interface {
	GetAllCategories() ([]*Category, error)
	GetCategoryByID(id int) (*Category, error)
	CreateCategory(req CreateCategoryRequest) (*Category, error)
	UpdateCategory(id int, req CreateCategoryRequest) (*Category, error)
	DeleteCategory(id int) error
}

type categoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}


func (s *categoryService) GetAllCategories() ([]*Category, error) {
	return s.repo.GetAll()
}

func (s *categoryService) GetCategoryByID(id int) (*Category, error) {
	return s.repo.GetByID(id)
}

func (s *categoryService) CreateCategory(req CreateCategoryRequest) (*Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name harus diisi")
	}

	category := &Category{
		Name:        req.Name,
		Description: req.Description,
	}

	err := s.repo.Create(category)
	if err != nil {
		return nil, fmt.Errorf("kategori gagal ditambahkan: %w", err)
	}

	return category, nil
}

func (s *categoryService) UpdateCategory(id int, req CreateCategoryRequest) (*Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name harus diisi")
	}

	category := &Category{
		Name:        req.Name,
		Description: req.Description,
	}

	err := s.repo.Update(id, category)
	if err != nil {
		return nil, fmt.Errorf("category gagal diperbarui: %w", err)
	}

	// amvil kategori yang diperbarui
	updatedCategory, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan kategori yang diperbarui: %w", err)
	}

	return updatedCategory, nil
}

func (s *categoryService) DeleteCategory(id int) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus kategori: %w", err)
	}
	return nil
}