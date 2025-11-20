package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoriesRepository struct {
	mock.Mock
}

func (m *MockCategoriesRepository) GetAll() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoriesRepository) GetByCode(code string) (*models.Category, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoriesRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func TestCategoriesHandler_HandleGet(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		categories := []models.Category{
			{Code: "CLOTHING", Name: "Clothing"},
			{Code: "SHOES", Name: "Shoes"},
			{Code: "ACCESSORIES", Name: "Accessories"},
		}

		mockRepo.On("GetAll").Return(categories, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response []CategoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 3)
		assert.Equal(t, "CLOTHING", response[0].Code)
		assert.Equal(t, "Clothing", response[0].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		mockRepo.On("GetAll").Return(nil, assert.AnError)

		req := httptest.NewRequest("GET", "/categories", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestCategoriesHandler_HandlePost(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "NEW_CATEGORY",
			Name: "New Category",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mockRepo.On("Create", mock.MatchedBy(func(cat *models.Category) bool {
			return cat.Code == "NEW_CATEGORY" && cat.Name == "New Category"
		})).Return(nil).Run(func(args mock.Arguments) {
			cat := args.Get(0).(*models.Category)
			cat.ID = 1
		})

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response CategoryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "NEW_CATEGORY", response.Code)
		assert.Equal(t, "New Category", response.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("missing code field", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Name: "New Category",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("missing name field", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "NEW_CATEGORY",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "NEW_CATEGORY",
			Name: "New Category",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mockRepo.On("Create", mock.Anything).Return(assert.AnError)

		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

