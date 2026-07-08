package handlers

import (
	"net/http"
	"strconv"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	auth *auth.Authenticator
}

func NewAPIKeyHandler(auth *auth.Authenticator) *APIKeyHandler {
	return &APIKeyHandler{auth: auth}
}

func (h *APIKeyHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "unauthorized"})
		return
	}

	uid, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user"})
		return
	}

	keys, err := h.auth.APIKeys().GetByUser(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	type maskedKey struct {
		ID        uint   `json:"id"`
		Provider  string `json:"provider"`
		Name      string `json:"name"`
		KeyPrefix string `json:"keyPrefix"`
		Status    string `json:"status"`
		CreatedAt string `json:"createdAt"`
	}

	result := make([]maskedKey, 0, len(keys))
	for _, k := range keys {
		result = append(result, maskedKey{
			ID:        k.ID,
			Provider:  k.Provider,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			Status:    k.Status,
			CreatedAt: k.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "unauthorized"})
		return
	}

	uid, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user"})
		return
	}

	var req struct {
		Provider string `json:"provider" binding:"required"`
		Name     string `json:"name"`
		Key      string `json:"key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	apiKey, err := h.auth.APIKeys().Create(auth.CreateAPIKeyParams{
		UserID:   uint(uid),
		Provider: req.Provider,
		Name:     req.Name,
		Key:      req.Key,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "API key saved",
		"data": gin.H{
			"id":        apiKey.ID,
			"provider":  apiKey.Provider,
			"name":      apiKey.Name,
			"keyPrefix": apiKey.KeyPrefix,
			"status":    apiKey.Status,
		},
	})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "unauthorized"})
		return
	}

	uid, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user"})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid key id"})
		return
	}

	if err := h.auth.APIKeys().Delete(uint(keyID), uint(uid)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "API key deleted"})
}

func (h *APIKeyHandler) GetProviders(c *gin.Context) {
	providers := h.auth.APIKeys().ProviderList()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": providers})
}
