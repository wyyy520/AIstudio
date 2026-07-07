package handlers

import (
	"net/http"
	"time"

	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains the JWT token on successful login.
type LoginResponse struct {
	Token    string `json:"token"`
	ExpiresIn int64 `json:"expires_in"` // seconds
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	// In production, this would reference a UserService for credential validation.
	// For now, we use a simple credential store.
	validUsers map[string]string // username -> password
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler() *AuthHandler {
	// Default development credentials; override in production via env/config.
	return &AuthHandler{
		validUsers: map[string]string{
			"admin": "admin123",
		},
	}
}

// Login authenticates a user and returns a JWT token.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request: username and password are required",
		})
		return
	}

	// Validate credentials
	expectedPassword, exists := h.validUsers[req.Username]
	if !exists || expectedPassword != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "invalid username or password",
		})
		return
	}

	// Generate JWT token (24h expiry)
	ttl := 24 * time.Hour
	token, err := middleware.GenerateToken(req.Username, req.Username, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "login successful",
		"data": LoginResponse{
			Token:    token,
			ExpiresIn: int64(ttl.Seconds()),
		},
	})
}