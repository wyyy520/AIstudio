package handlers

import (
	"net/http"
	"strconv"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	auth *auth.Authenticator
}

func NewUserHandler(auth *auth.Authenticator) *UserHandler {
	return &UserHandler{auth: auth}
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.auth.Users().List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": users})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user id"})
		return
	}

	user, err := h.auth.Users().GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": user})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	user, err := h.auth.Users().Create(req.Username, req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err == auth.ErrDuplicateUser {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"code": -1, "message": err.Error()})
		return
	}

	h.auth.Quotas().InitDefaults(user.ID)

	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "created", "data": user})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user id"})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Status   string `json:"status"`
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
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	user, err := h.auth.Users().Update(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": user})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid user id"})
		return
	}

	if err := h.auth.Users().Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

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

	if err := h.auth.Users().ChangePassword(uint(uid), req.OldPassword, req.NewPassword); err != nil {
		status := http.StatusBadRequest
		if err == auth.ErrUserNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "password changed"})
}
