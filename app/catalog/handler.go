package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

type Product struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category *CategoryDTO `json:"category,omitempty"`
}

type CategoryDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CatalogHandler struct {
	repo models.ProductsRepository
}

func NewCatalogHandler(r models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	filters := models.ProductFilters{}

	// Parse category filter
	if categoryCode := r.URL.Query().Get("category"); categoryCode != "" {
		filters.CategoryCode = &categoryCode
	}

	// Parse price filter
	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid price_less_than parameter")
			return
		}
		filters.PriceLessThan = &price
	}

	// Parse pagination parameters
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid limit parameter (must be between 1 and 100)")
			return
		}
	}

	// Fetch products
	products, total, err := h.repo.GetAll(filters, offset, limit)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	productDTOs := make([]Product, len(products))
	for i, p := range products {
		productDTOs[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
		}
		if p.Category != nil {
			productDTOs[i].Category = &CategoryDTO{
				Code: p.Category.Code,
				Name: p.Category.Name,
			}
		}
	}

	response := Response{
		Products: productDTOs,
		Total:    total,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.repo.GetByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}

	// Map variants with inherited prices
	variants := make([]VariantDTO, len(product.Variants))
	for i, v := range product.Variants {
		price := product.Price
		if v.Price.IsPositive() {
			price = v.Price
		}
		variants[i] = VariantDTO{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	productDetail := ProductDetail{
		Code:     product.Code,
		Price:    product.Price.InexactFloat64(),
		Variants: variants,
	}

	if product.Category != nil {
		productDetail.Category = &CategoryDTO{
			Code: product.Category.Code,
			Name: product.Category.Name,
		}
	}

	api.OKResponse(w, productDetail)
}

type ProductDetail struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category *CategoryDTO `json:"category,omitempty"`
	Variants []VariantDTO `json:"variants"`
}

type VariantDTO struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}
