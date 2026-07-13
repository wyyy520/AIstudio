package handlers

import (
	"net/http"
	"strconv"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	auth *auth.Authenticator
}

func NewProfileHandler(auth *auth.Authenticator) *ProfileHandler {
	return &ProfileHandler{auth: auth}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
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

	user, err := h.auth.Users().GetByID(uint(uid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"nickname":  user.Nickname,
			"avatar":    user.Avatar,
			"role":      user.Role,
			"status":    user.Status,
			"createdAt": user.CreatedAt,
		},
	})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
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
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	user, err := h.auth.Users().Update(uint(uid), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "profile updated",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
		},
	})
}

func (h *ProfileHandler) GetSessions(c *gin.Context) {
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

	sessions, err := h.auth.Sessions().GetByUserID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": sessions})
}
