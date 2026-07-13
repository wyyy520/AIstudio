package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type AuthHandler struct {
	auth *auth.Authenticator
}

func NewAuthHandler(auth *auth.Authenticator) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request: username and password are required",
		})
		return
	}

	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	result, err := h.auth.Login(auth.LoginParams{
		Username:   req.Username,
		Password:   req.Password,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
	})
	if err != nil {
		status := http.StatusUnauthorized
		if err == auth.ErrUserDisabled {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "login successful",
		"data":    result,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if len(token) > 7 {
		token = token[7:]
	}

	if err := h.auth.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "logout successful",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	opts := []auth.UserOption{}
	if req.Nickname != "" {
		opts = append(opts, auth.WithNickname(req.Nickname))
	}

	user, err := h.auth.Users().Create(req.Username, req.Email, req.Password, opts...)
	if err != nil {
		status := http.StatusInternalServerError
		if err == auth.ErrDuplicateUser {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	h.auth.Quotas().InitDefaults(user.ID)

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "registration successful",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "refreshToken is required",
		})
		return
	}

	accessToken, err := h.auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "token refreshed",
		"data": gin.H{
			"accessToken": accessToken,
		},
	})
}
