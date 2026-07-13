package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	setupTestEnvironment()

	// test registration
	registerBody := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
		"email":    "test@example.com",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("register expected 200/201, got %d: %s", w.Code, w.Body.String())
	}

	// test login returns JWT
	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var loginResp struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if loginResp.Code != 0 {
		t.Fatalf("login response code expected 0, got %d: %s", loginResp.Code, w.Body.String())
	}

	accessToken, ok := loginResp.Data["accessToken"].(string)
	if !ok || accessToken == "" {
		t.Fatal("login response missing accessToken")
	}

	// test protected endpoint with JWT
	req = httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("profile endpoint expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// test protected endpoint without JWT returns 401
	req = httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("expected 401 for unauthenticated request, got 200")
	}
}

func TestAuthFlowProjectCRUD(t *testing.T) {
	setupTestEnvironment()

	loginBody := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
	}
	body, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Skipf("login failed, skipping project CRUD test: %d", w.Code)
	}

	var loginResp struct {
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	accessToken, _ := loginResp.Data["accessToken"].(string)

	projectBody := map[string]interface{}{
		"name":        "test-project",
		"description": "e2e test project",
	}
	body, _ = json.Marshal(projectBody)
	req = httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("create project expected 200/201, got %d: %s", w.Code, w.Body.String())
	}
}
