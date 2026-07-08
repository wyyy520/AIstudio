package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/database"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
	"github.com/gin-gonic/gin"
)

type testResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    json.RawMessage  `json:"data,omitempty"`
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	config.Load()

	// Use in-memory SQLite
	cfg, _ := database.LoadConfig()
	database.Init(cfg)

	// Task manager
	taskMgr := task.NewManager(2)
	taskMgr.Start()

	// Plugin manager
	pluginMgr := plugin.NewManager("../../../../Plugins")
	pluginMgr.DiscoverPlugins()

	// Workflow engine
	workflow.RegisterDefaultNodes()
	engine := workflow.NewDefaultEngine()

	// Register workflow/agent task handlers
	workflowTaskHandler := workflow.NewTaskHandler(engine)
	workflowTaskHandler.SetTaskManager(taskMgr)
	taskMgr.RegisterHandler("workflow", workflowTaskHandler)
	taskMgr.RegisterHandler("agent", workflowTaskHandler)

	// Environment manager
	envMgr := environment.NewManager()

	// Services
	var llmProvider agent.LLMProvider
	agentMemory, _ := agent.NewMemory(database.GetDB())
	aiAgent := agent.NewAgent(llmProvider, agentMemory)
	svc := service.NewServices(database.GetDB(), taskMgr, pluginMgr, engine, envMgr, aiAgent)

	r := gin.New()
	r.Use(gin.Recovery())

	// Register routes
	projectHandler := NewProjectHandler(svc.Project)
	projects := r.Group("/api/projects")
	{
		projects.GET("", projectHandler.List)
		projects.GET("/:id", projectHandler.Get)
		projects.POST("", projectHandler.Create)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
	}

	taskHandler := NewTaskHandler(svc.Task)
	tasks := r.Group("/api/tasks")
	{
		tasks.GET("", taskHandler.List)
		tasks.GET("/:id", taskHandler.Get)
		tasks.POST("", taskHandler.Create)
	}

	pluginHandler := NewPluginHandler(svc.Plugin)
	plugins := r.Group("/api/plugins")
	{
		plugins.GET("", pluginHandler.List)
		plugins.GET("/:name", pluginHandler.Get)
	}

	agentHandler := NewAgentHandler(svc.Agent)
	agent := r.Group("/api/agent")
	{
		agent.POST("/chat", agentHandler.Chat)
	}

	logHandler := NewLogHandler(svc.Log)
	logs := r.Group("/api/logs")
	{
		logs.GET("", logHandler.Query)
	}

	envHandler := NewEnvironmentHandler(svc.Environment)
	env := r.Group("/api/environment")
	{
		env.GET("/status", envHandler.GetStatus)
	}

	return r
}

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestProjectCRUD(t *testing.T) {
	router := setupTestRouter()

	// Create
	w := performRequest(router, "POST", "/api/projects", `{"name":"Test Project","description":"A test project","ownerId":1}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d: %s", resp.Code, string(resp.Data))
	}

	// Extract project ID
	var project struct {
		ID uint `json:"id"`
	}
	json.Unmarshal(resp.Data, &project)
	if project.ID == 0 {
		t.Fatal("expected non-zero project ID")
	}
	projectID := project.ID

	// List
	w = performRequest(router, "GET", "/api/projects", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Get by ID
	w = performRequest(router, "GET", fmt.Sprintf("/api/projects/%d", projectID), "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Update
	w = performRequest(router, "PUT", fmt.Sprintf("/api/projects/%d", projectID), `{"name":"Updated Project"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete
	w = performRequest(router, "DELETE", fmt.Sprintf("/api/projects/%d", projectID), "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTaskCreate(t *testing.T) {
	router := setupTestRouter()

	// Create a task (requires a registered handler)
	w := performRequest(router, "POST", "/api/tasks", `{"name":"Test Task","handler":"workflow","priority":1}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}

	// Invalid handler
	w = performRequest(router, "POST", "/api/tasks", `{"name":"Bad Task","handler":"nonexistent","priority":1}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid handler, got %d", w.Code)
	}

	// List tasks
	w = performRequest(router, "GET", "/api/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPluginQuery(t *testing.T) {
	router := setupTestRouter()

	// List plugins (may be empty if test environment doesn't have plugins)
	w := performRequest(router, "GET", "/api/plugins", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestAgentChat(t *testing.T) {
	router := setupTestRouter()

	// Chat with run intent
	w := performRequest(router, "POST", "/api/agent/chat", `{"message":"run the workflow","projectId":"1","context":{}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}

	// Chat with help intent
	w = performRequest(router, "POST", "/api/agent/chat", `{"message":"help","projectId":"1","context":{}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogQuery(t *testing.T) {
	router := setupTestRouter()

	// Query logs with default pagination
	w := performRequest(router, "GET", "/api/logs?page=1&size=10", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestChatRequest(t *testing.T) {
	// Verify the ChatRequest JSON structure
	svc := &service.AgentService{}
	_ = svc // just checking it compiles

	req := service.ChatRequest{
		Message:   "test",
		ProjectID: "1",
		Context:   map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal ChatRequest failed: %v", err)
	}

	var unmarshaled service.ChatRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal ChatRequest failed: %v", err)
	}

	if unmarshaled.Message != "test" {
		t.Errorf("expected 'test', got '%s'", unmarshaled.Message)
	}
}

func TestEnvironmentStatus(t *testing.T) {
	router := setupTestRouter()

	w := performRequest(router, "GET", "/api/environment/status", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d: %s", resp.Code, string(resp.Data))
	}

	var status environment.EnvironmentStatus
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		t.Fatalf("unmarshal EnvironmentStatus failed: %v", err)
	}

	t.Logf("Python: %+v", status.Python)
	t.Logf("CUDA: %+v", status.CUDA)
	t.Logf("Dependencies: %+v", status.Dependencies)
}