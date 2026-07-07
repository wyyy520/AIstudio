package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestEngine creates a Gin engine with all middleware for testing.
func setupTestEngine(cfg Config) *gin.Engine {
	r := gin.New()
	Apply(r, cfg)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/projects", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "projects"})
	})
	r.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "login"})
	})

	return r
}

// ============================================================
// 1. Auth Middleware Tests
// ============================================================

func TestAuth_PublicPath(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("public path /api/health expected 200, got %d", w.Code)
	}
}

func TestAuth_LoginPath(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	r.ServeHTTP(w, req)

	// Login should be accessible without auth (public)
	if w.Code == http.StatusUnauthorized {
		t.Fatal("login path should be public, got 401")
	}
}

func TestAuth_MissingToken(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("private path without token expected 401, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "missing authorization header" {
		t.Fatalf("expected 'missing authorization header', got '%v'", resp["message"])
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token expected 401, got %d", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	// Generate a valid token
	token, err := GenerateToken("user123", "testuser", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("valid token expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	// Generate an already-expired token
	token, err := GenerateToken("user123", "testuser", -time.Second)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expired token expected 401, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	msg, _ := resp["message"].(string)
	if !strings.Contains(msg, "expired") {
		t.Fatalf("expected 'expired' message, got '%s'", msg)
	}
}

func TestAuth_WrongFormat(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Basic somecredentials")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong auth format expected 401, got %d", w.Code)
	}
}

// ============================================================
// 2. CORS Middleware Tests
// ============================================================

func TestCORS_AllowedOrigin(t *testing.T) {
	r := setupTestEngine(Config{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Fatalf("expected origin 'http://localhost:5173', got '%s'", origin)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
	}
}

func TestCORS_Preflight(t *testing.T) {
	r := setupTestEngine(Config{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/projects", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight expected 204, got %d", w.Code)
	}

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Fatal("preflight response should include Allow-Origin header")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CORS = &CORSConfig{
		AllowedOrigins: []string{"http://trusted.com"},
	}
	r := setupTestEngine(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Fatalf("disallowed origin should not have CORS header, got '%s'", origin)
	}
}

// ============================================================
// 3. Logger Middleware Tests
// ============================================================

func TestLogger_StructuredOutput(t *testing.T) {
	// Capture stdout
	var buf strings.Builder
	oldOutput := loggerOutput
	defer func() { loggerOutput = oldOutput }()

	tmpFile, _ := os.CreateTemp("", "logger-test-*")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	SetLoggerOutput(tmpFile)

	r := setupTestEngine(Config{JWTSecret: "test"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	// Read the log output
	tmpFile.Seek(0, 0)
	content := make([]byte, 1024)
	n, _ := tmpFile.Read(content)
	logLine := string(content[:n])

	if !strings.Contains(logLine, `"method":"GET"`) {
		t.Fatalf("logger output missing method: %s", logLine)
	}
	if !strings.Contains(logLine, `"path":"/api/health"`) {
		t.Fatalf("logger output missing path: %s", logLine)
	}
	if !strings.Contains(logLine, `"status":200`) {
		t.Fatalf("logger output missing status: %s", logLine)
	}

	_ = buf
}

// ============================================================
// 4. Recovery Middleware Tests
// ============================================================

func TestRecovery_Panic(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test"})

	AddPublicPath("/api/panic")

	// Add a handler that panics
	r.GET("/api/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("panic handler expected 500, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != -1.0 {
		t.Fatalf("expected code -1, got %v", resp["code"])
	}
}

func TestRecovery_PanicWithDevelopment(t *testing.T) {
	r := setupTestEngine(Config{
		JWTSecret:   "test",
		Development: true,
	})

	AddPublicPath("/api/panic-dev")

	r.GET("/api/panic-dev", func(c *gin.Context) {
		panic("dev panic detail")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/panic-dev", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	msg, _ := resp["message"].(string)
	if !strings.Contains(msg, "dev panic detail") {
		t.Fatalf("dev mode should include panic detail, got '%s'", msg)
	}
}

// ============================================================
// 5. Rate Limit Middleware Tests
// ============================================================

func TestRateLimit_Allowed(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test"})

	// First request should be allowed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", w.Code)
	}
}

func TestRateLimit_Exceeded(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RateLimit = &RateLimitConfig{
		Rate:  10.0, // 10 tokens/sec
		Burst: 10,   // burst of 10
	}
	cfg.JWTSecret = "test"
	r := setupTestEngine(cfg)

	AddPublicPath("/api/ratelimit-test")

	// Add a handler for rate limit testing
	r.GET("/api/ratelimit-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Consume all burst tokens
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/ratelimit-test", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d expected 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/ratelimit-test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("rate-limited request expected 429, got %d", w.Code)
	}
}

func TestRateLimit_PreflightSkipped(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test"})

	// Preflight OPTIONS should bypass rate limiting
	for i := 0; i < 200; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/api/projects", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("preflight %d expected 204, got %d", i+1, w.Code)
		}
	}
}

// ============================================================
// 6. JWT Utility Tests
// ============================================================

func TestJWT_GenerateAndValidate(t *testing.T) {
	ResetJWTSecret()
	SetJWTSecret("test-secret")

	token, err := GenerateToken("42", "alice", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Token should have 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}

	// Validate
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Sub != "42" {
		t.Fatalf("expected sub '42', got '%s'", claims.Sub)
	}
	if claims.Username != "alice" {
		t.Fatalf("expected username 'alice', got '%s'", claims.Username)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	ResetJWTSecret()
	SetJWTSecret("test-secret")

	// Generate token that expires immediately
	token, err := GenerateToken("1", "test", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token)
	if err == nil {
		t.Fatal("expected validation error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected 'expired' error, got '%v'", err)
	}
}

func TestJWT_InvalidSignature(t *testing.T) {
	ResetJWTSecret()
	SetJWTSecret("secret1")

	token, err := GenerateToken("1", "test", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with different secret
	ResetJWTSecret()
	SetJWTSecret("secret2")
	_, err = ValidateToken(token)
	if err == nil {
		t.Fatal("expected validation error for wrong signature")
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	ResetJWTSecret()
	SetJWTSecret("test-secret")

	_, err := ValidateToken("not-a-valid-jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}

	_, err = ValidateToken("part1.part2")
	if err == nil {
		t.Fatal("expected error for 2-part token")
	}
}

// ============================================================
// 7. Middleware Config Tests
// ============================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.JWTSecret != "" {
		t.Fatalf("expected empty JWT secret, got '%s'", cfg.JWTSecret)
	}
	if cfg.Development {
		t.Fatal("expected Development=false")
	}
}

// ============================================================
// 8. Integration: Full Middleware Stack
// ============================================================

func TestFullStack_MissingAuth(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test"})

	// Private route without auth should be rejected before reaching handler
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	// Should also have CORS headers
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Fatal("response should include CORS headers even on auth failure")
	}
}

func TestFullStack_Authenticated(t *testing.T) {
	r := setupTestEngine(Config{JWTSecret: "test"})

	token, _ := GenerateToken("1", "testuser", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("authenticated request expected 200, got %d", w.Code)
	}

	// Should have CORS headers
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Fatal("authenticated response should include CORS headers")
	}
}