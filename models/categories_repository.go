package models

import (
	"gorm.io/gorm"
)

type CategoriesRepository interface {
	GetAll() ([]Category, error)
	GetByCode(code string) (*Category, error)
	Create(category *Category) error
}

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepository {
	return &categoriesRepository{
		db: db,
	}
}

func (r *categoriesRepository) GetAll() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoriesRepository) GetByCode(code string) (*Category, error) {
	var category Category
	if err := r.db.Where("code = ?", code).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoriesRepository) Create(category *Category) error {
	return r.db.Create(category).Error
}

