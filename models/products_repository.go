package models

import (
	"gorm.io/gorm"
)

type ProductsRepository interface {
	GetAll(filters ProductFilters, offset, limit int) ([]Product, int64, error)
	GetByCode(code string) (*Product, error)
}

type ProductFilters struct {
	CategoryCode *string
	PriceLessThan *float64
}

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) GetAll(filters ProductFilters, offset, limit int) ([]Product, int64, error) {
	baseQuery := r.db.Model(&Product{})
	countQuery := r.db.Model(&Product{})

	if filters.CategoryCode != nil {
		joinClause := "JOIN categories ON products.category_id = categories.id"
		baseQuery = baseQuery.Joins(joinClause).Where("categories.code = ?", *filters.CategoryCode)
		countQuery = countQuery.Joins(joinClause).Where("categories.code = ?", *filters.CategoryCode)
	}

	if filters.PriceLessThan != nil {
		baseQuery = baseQuery.Where("products.price < ?", *filters.PriceLessThan)
		countQuery = countQuery.Where("products.price < ?", *filters.PriceLessThan)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []Product
	if err := baseQuery.Preload("Category").Preload("Variants").
		Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productsRepository) GetByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").
		Where("code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
