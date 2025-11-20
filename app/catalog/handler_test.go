package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductsRepository struct {
	mock.Mock
}

func (m *MockProductsRepository) GetAll(filters models.ProductFilters, offset, limit int) ([]models.Product, int64, error) {
	args := m.Called(filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductsRepository) GetByCode(code string) (*models.Product, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestCatalogHandler_HandleGet(t *testing.T) {
	t.Run("successful request with default pagination", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{
			{Code: "PROD001", Price: decimal.NewFromFloat(10.99)},
			{Code: "PROD002", Price: decimal.NewFromFloat(12.49)},
		}

		mockRepo.On("GetAll", models.ProductFilters{}, 0, 10).Return(products, int64(2), nil)

		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Len(t, response.Products, 2)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, 10.99, response.Products[0].Price)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful request with pagination parameters", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{
			{Code: "PROD003", Price: decimal.NewFromFloat(8.75)},
		}

		mockRepo.On("GetAll", models.ProductFilters{}, 5, 20).Return(products, int64(8), nil)

		req := httptest.NewRequest("GET", "/catalog?offset=5&limit=20", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(8), response.Total)
		assert.Len(t, response.Products, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful request with category filter", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		categoryCode := "CLOTHING"
		category := &models.Category{Code: "CLOTHING", Name: "Clothing"}
		products := []models.Product{
			{Code: "PROD001", Price: decimal.NewFromFloat(10.99), Category: category},
		}

		filters := models.ProductFilters{CategoryCode: &categoryCode}
		mockRepo.On("GetAll", filters, 0, 10).Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog?category=CLOTHING", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Products, 1)
		assert.NotNil(t, response.Products[0].Category)
		assert.Equal(t, "CLOTHING", response.Products[0].Category.Code)
		assert.Equal(t, "Clothing", response.Products[0].Category.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful request with price filter", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		price := 15.0
		products := []models.Product{
			{Code: "PROD001", Price: decimal.NewFromFloat(10.99)},
		}

		filters := models.ProductFilters{PriceLessThan: &price}
		mockRepo.On("GetAll", filters, 0, 10).Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog?price_less_than=15", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid offset parameter", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?offset=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "invalid offset")

		mockRepo.AssertNotCalled(t, "GetAll")
	})

	t.Run("invalid limit parameter - too low", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?limit=0", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "GetAll")
	})

	t.Run("invalid limit parameter - too high", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?limit=101", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "GetAll")
	})

	t.Run("invalid price_less_than parameter", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?price_less_than=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "GetAll")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		mockRepo.On("GetAll", models.ProductFilters{}, 0, 10).Return(nil, int64(0), assert.AnError)

		req := httptest.NewRequest("GET", "/catalog", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestCatalogHandler_HandleGetByCode(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		category := &models.Category{Code: "CLOTHING", Name: "Clothing"}
		product := &models.Product{
			Code:     "PROD001",
			Price:    decimal.NewFromFloat(10.99),
			Category: category,
			Variants: []models.Variant{
				{Name: "Variant A", SKU: "SKU001A", Price: decimal.NewFromFloat(11.99)},
				{Name: "Variant B", SKU: "SKU001B", Price: decimal.Zero},
			},
		}

		mockRepo.On("GetByCode", "PROD001").Return(product, nil)

		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ProductDetail
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PROD001", response.Code)
		assert.Equal(t, 10.99, response.Price)
		assert.NotNil(t, response.Category)
		assert.Equal(t, "CLOTHING", response.Category.Code)
		assert.Len(t, response.Variants, 2)
		assert.Equal(t, 11.99, response.Variants[0].Price)
		assert.Equal(t, 10.99, response.Variants[1].Price) // inherited from product

		mockRepo.AssertExpectations(t)
	})

	t.Run("product not found", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		mockRepo.On("GetByCode", "PROD999").Return(nil, assert.AnError)

		req := httptest.NewRequest("GET", "/catalog/PROD999", nil)
		req.SetPathValue("code", "PROD999")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("missing code parameter", func(t *testing.T) {
		mockRepo := new(MockProductsRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog/", nil)
		req.SetPathValue("code", "")
		w := httptest.NewRecorder()

		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "GetByCode")
	})
}

