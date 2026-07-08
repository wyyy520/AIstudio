package handlers

import (
	"net/http"
	"strconv"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type QuotaHandler struct {
	auth *auth.Authenticator
}

func NewQuotaHandler(auth *auth.Authenticator) *QuotaHandler {
	return &QuotaHandler{auth: auth}
}

func (h *QuotaHandler) GetQuotas(c *gin.Context) {
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

	quotas, err := h.auth.Quotas().GetUserQuotas(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	type quotaInfo struct {
		Resource string `json:"resource"`
		Limit    int64  `json:"limit"`
		Used     int64  `json:"used"`
		Remaining int64 `json:"remaining"`
	}

	result := make([]quotaInfo, 0, len(quotas))
	for _, q := range quotas {
		remaining := q.Limit - q.Used
		if q.Limit < 0 {
			remaining = -1
		}
		result = append(result, quotaInfo{
			Resource:  q.ResourceType,
			Limit:     q.Limit,
			Used:      q.Used,
			Remaining: remaining,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

func (h *QuotaHandler) UpdateQuota(c *gin.Context) {
	var req struct {
		UserID       uint   `json:"userId" binding:"required"`
		ResourceType string `json:"resourceType" binding:"required"`
		Limit        int64  `json:"limit" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if err := h.auth.Quotas().UpdateLimit(req.UserID, auth.QuotaResource(req.ResourceType), req.Limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "quota updated"})
}

func (h *QuotaHandler) CheckQuota(c *gin.Context) {
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

	resource := c.Query("resource")
	if resource == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "resource parameter is required"})
		return
	}

	if err := h.auth.Quotas().CheckQuota(uint(uid), auth.QuotaResource(resource)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
			"data": gin.H{
				"allowed": false,
				"reason":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "quota available",
		"data": gin.H{
			"allowed": true,
		},
	})
}
