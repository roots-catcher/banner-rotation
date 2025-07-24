package api

import (
	"banner-rotation/internal/app"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBandit реализует app.BanditInterface для тестов
type MockBandit struct {
	mock.Mock
}

func (m *MockBandit) AddBannerToSlot(ctx context.Context, slotID, bannerID int) error {
	args := m.Called(ctx, slotID, bannerID)
	return args.Error(0)
}

func (m *MockBandit) RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error {
	args := m.Called(ctx, slotID, bannerID)
	return args.Error(0)
}

func (m *MockBandit) ChooseBanner(ctx context.Context, slotID, groupID int) (int, error) {
	args := m.Called(ctx, slotID, groupID)
	return args.Int(0), args.Error(1)
}

func (m *MockBandit) RecordClick(ctx context.Context, slotID, bannerID, groupID int) error {
	args := m.Called(ctx, slotID, bannerID, groupID)
	return args.Error(0)
}

func TestAPIEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем мок, реализующий интерфейс BanditInterface
	mockBandit := new(MockBandit)
	server := NewServer(mockBandit)

	t.Run("AddBannerToSlot - success", func(t *testing.T) {
		mockBandit.On("AddBannerToSlot", mock.Anything, 1, 100).Return(nil)

		w := httptest.NewRecorder()
		req := createRequest(t, "POST", "/api/v1/banner_slot", AddBannerToSlotRequest{
			SlotID:   1,
			BannerID: 100,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		mockBandit.AssertExpectations(t)
	})

	t.Run("AddBannerToSlot - invalid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := createRequest(t, "POST", "/api/v1/banner_slot", map[string]interface{}{
			"slot_id":   "invalid", // Неправильный тип
			"banner_id": 100,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("RemoveBannerFromSlot - success", func(t *testing.T) {
		mockBandit.On("RemoveBannerFromSlot", mock.Anything, 1, 100).Return(nil)

		w := httptest.NewRecorder()
		req := createRequest(t, "DELETE", "/api/v1/banner_slot", RemoveBannerFromSlotRequest{
			SlotID:   1,
			BannerID: 100,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		mockBandit.AssertExpectations(t)
	})

	t.Run("ChooseBanner - success", func(t *testing.T) {
		mockBandit.On("ChooseBanner", mock.Anything, 1, 1).Return(100, nil)

		w := httptest.NewRecorder()
		req := createRequest(t, "POST", "/api/v1/choose_banner", ChooseBannerRequest{
			SlotID:  1,
			GroupID: 1,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp ChooseBannerResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 100, resp.BannerID)
		mockBandit.AssertExpectations(t)
	})

	t.Run("ChooseBanner - no banners", func(t *testing.T) {
		mockBandit.On("ChooseBanner", mock.Anything, 2, 1).Return(0, app.ErrNoBanners)

		w := httptest.NewRecorder()
		req := createRequest(t, "POST", "/api/v1/choose_banner", ChooseBannerRequest{
			SlotID:  2,
			GroupID: 1,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockBandit.AssertExpectations(t)
	})

	t.Run("RegisterClick - success", func(t *testing.T) {
		mockBandit.On("RecordClick", mock.Anything, 1, 100, 1).Return(nil)

		w := httptest.NewRecorder()
		req := createRequest(t, "POST", "/api/v1/register_click", RegisterClickRequest{
			SlotID:   1,
			BannerID: 100,
			GroupID:  1,
		})

		server.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		mockBandit.AssertExpectations(t)
	})
}

func createRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req
}
