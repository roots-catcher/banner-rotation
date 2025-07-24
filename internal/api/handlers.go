package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddBannerToSlotRequest запрос на добавление баннера в слот
type AddBannerToSlotRequest struct {
	SlotID   int `json:"slot_id" binding:"required"`
	BannerID int `json:"banner_id" binding:"required"`
}

// RemoveBannerFromSlotRequest запрос на удаление баннера из слота
type RemoveBannerFromSlotRequest struct {
	SlotID   int `json:"slot_id" binding:"required"`
	BannerID int `json:"banner_id" binding:"required"`
}

// ChooseBannerRequest запрос на выбор баннера
type ChooseBannerRequest struct {
	SlotID  int `json:"slot_id" binding:"required"`
	GroupID int `json:"group_id" binding:"required"`
}

// ChooseBannerResponse ответ с выбранным баннером
type ChooseBannerResponse struct {
	BannerID int `json:"banner_id"`
}

// RegisterClickRequest запрос на регистрацию клика
type RegisterClickRequest struct {
	SlotID   int `json:"slot_id" binding:"required"`
	BannerID int `json:"banner_id" binding:"required"`
	GroupID  int `json:"group_id" binding:"required"`
}

func (s *Server) addBannerToSlot(c *gin.Context) {
	var req AddBannerToSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.bandit.AddBannerToSlot(c.Request.Context(), req.SlotID, req.BannerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) removeBannerFromSlot(c *gin.Context) {
	var req RemoveBannerFromSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.bandit.RemoveBannerFromSlot(c.Request.Context(), req.SlotID, req.BannerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) chooseBanner(c *gin.Context) {
	var req ChooseBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bannerID, err := s.bandit.ChooseBanner(c.Request.Context(), req.SlotID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChooseBannerResponse{BannerID: bannerID})
}

func (s *Server) registerClick(c *gin.Context) {
	var req RegisterClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.bandit.RecordClick(c.Request.Context(), req.SlotID, req.BannerID, req.GroupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
