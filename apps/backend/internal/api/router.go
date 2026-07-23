// Package api defines the HTTP API routing layer for AIStudio.
//
// This file sets up ALL API routes using the Gin framework. Routes are
// organized into functional groups:
//
//	Health        GET  /api/health
//	Auth          POST /api/auth/login, /register, /refresh, /logout
//	User Profile  GET/PUT /api/user/profile, /sessions
//	API Keys      GET/POST/DELETE /api/user/apikeys
//	Quota         GET /api/user/quota, POST /api/admin/quota
//	Users         CRUD /api/users
//	Projects      CRUD + Workflow + Compile + Run /api/projects
//	Runtime       POST /api/runtime/detect, /list, /status, /stop
//	Workflows     CRUD + Run /api/workflows
//	Logs          GET /api/logs
//	Tasks         CRUD + Cancel /api/tasks
//	Plugins       CRUD + Install + Execute /api/plugins
//	Agent         POST /api/agent/chat, /plan, /generate-workflow, /generate-and-run
//	MCP           CRUD /api/mcp/tools, /servers, /connect, /call
//	Environment   /api/environment/status, /check, /repair, /install
//	Error         POST /api/error/analyze, /repair
//	Settings      GET/PUT /api/settings, /engine
//	WebSocket     GET /api/ws, /api/v1/ws
//
// Architecture:
//
//	HTTP Request → Gin Router → Middleware (CORS, Logger, Recovery, Auth, RateLimit)
//	→ Handler → Service Layer → Module (Compiler, Runtime, Agent, ...)
//
// EngStudio.md §10 — API Reference
package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/aistudio/backend/internal/api/handlers"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/api/ws"
	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ============================================================================
// WebSocket Upgrader — origin validation for CORS in WebSocket handshake
// ============================================================================

func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			for _, allowed := range allowedOrigins {
				if allowed == "*" || allowed == origin {
					return true
				}
				if strings.HasPrefix(allowed, "*.") && strings.HasSuffix(origin, allowed[1:]) {
					return true
				}
			}
			return false
		},
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
}

// ============================================================================
// SetupRouter — the single entry point for all API routing
// Registers ~60+ endpoints across 19 handler groups.
// ============================================================================

// SetupRouter creates and configures the Gin router with all API endpoints.
// It is called once at startup from main.go's buildHandler().
func SetupRouter(svc *service.Services, mwCfg middleware.Config, wsHub *ws.Hub) *gin.Engine {
	r := gin.New()

	// Step 1: Apply global middleware (CORS, Logger, Recovery, RateLimit)
	middleware.Apply(r, mwCfg)

	// Step 2: Health — unauthenticated liveness check
	// ---- Health ----
	healthHandler := handlers.NewHealthHandler(svc)
	r.GET("/api/health", healthHandler.Check)

	// Step 3: Auth — public endpoints (login, register, refresh)
	// ---- Auth (public) ----
	authHandler := handlers.NewAuthHandler(svc.Authenticator)
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/refresh", authHandler.RefreshToken)

	// Step 4: Auth — authenticated (logout)
	// ---- Auth (authenticated) ----
	authGroup := r.Group("/api/auth")
	authGroup.POST("/logout", authHandler.Logout)

	// Step 5: User Profile — get/update profile and sessions
	// ---- User Profile ----
	profileHandler := handlers.NewProfileHandler(svc.Authenticator)
	profile := r.Group("/api/user")
	{
		profile.GET("/profile", profileHandler.GetProfile)
		profile.PUT("/profile", profileHandler.UpdateProfile)
		profile.GET("/sessions", profileHandler.GetSessions)
	}

	// Step 6: API Keys — manage LLM provider keys
	// ---- API Keys ----
	apiKeyHandler := handlers.NewAPIKeyHandler(svc.Authenticator)
	apiKeys := r.Group("/api/user/apikeys")
	{
		apiKeys.GET("", apiKeyHandler.List)
		apiKeys.POST("", apiKeyHandler.Create)
		apiKeys.DELETE("/:id", apiKeyHandler.Delete)
	}
	r.GET("/api/providers", apiKeyHandler.GetProviders)

	// Step 7: Quota — usage tracking and limits
	// ---- Quota ----
	quotaHandler := handlers.NewQuotaHandler(svc.Authenticator)
	quota := r.Group("/api/user/quota")
	{
		quota.GET("", quotaHandler.GetQuotas)
		quota.GET("/check", quotaHandler.CheckQuota)
	}

	// ---- Admin: Quota Management ----
	r.POST("/api/admin/quota", quotaHandler.UpdateQuota)

	// Step 8: Users — admin user management
	// ---- Users ----
	userHandler := handlers.NewUserHandler(svc.Authenticator)
	users := r.Group("/api/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.POST("", userHandler.Create)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
		users.PUT("/:id/password", userHandler.ChangePassword)
	}

	// Step 9: Projects — CRUD + workflow I/O + compile/run
	// ---- Projects (filesystem-based) ----
	projectHandler := handlers.NewProjectHandler(svc.ProjectManager)
	projects := r.Group("/api/projects")
	{
		projects.GET("", projectHandler.List)
		projects.GET("/recent", projectHandler.Recent)
		projects.GET("/:id", projectHandler.Get)
		projects.POST("", projectHandler.Create)
		projects.POST("/open", projectHandler.Open)
		projects.POST("/scan", projectHandler.Scan)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)

		projects.GET("/:id/workflow", projectHandler.ReadWorkflow)
		projects.PUT("/:id/workflow", projectHandler.SaveWorkflow)
	}

	// Step 10: Runtime — detect, execute, status, stop
	// ---- Runtime Pipeline (Compile → Generate → Execute) ----
	runtimeHandler := handlers.NewRuntimeHandler(svc.RuntimeService(), svc.Compiler)
	runtimeGroup := r.Group("/api/runtime")
	{
		runtimeGroup.POST("/detect", runtimeHandler.Detect)
		runtimeGroup.GET("/list", runtimeHandler.ListRunning)
		runtimeGroup.GET("/status/:runId", runtimeHandler.Status)
		runtimeGroup.POST("/stop", runtimeHandler.Stop)
	}
	// Per-project compile/run endpoints
	projects.POST("/:id/compile", runtimeHandler.Compile)
	projects.POST("/:id/run", runtimeHandler.Run)

	// Step 11: Workflows — CRUD + run + node type listing
	// ---- Workflows ----
	workflowHandler := handlers.NewWorkflowHandler(svc.WorkflowService())
	workflowHandler.SetTaskService(svc.TaskService())
	workflows := r.Group("/api/workflows")
	{
		workflows.GET("", workflowHandler.List)
		workflows.GET("/:id", workflowHandler.Get)
		workflows.POST("", workflowHandler.Create)
		workflows.PUT("/:id", workflowHandler.Update)
		workflows.DELETE("/:id", workflowHandler.Delete)
		workflows.POST("/:id/run", workflowHandler.Run)
		workflows.GET("/nodes", workflowHandler.ListNodeTypes)
	}

	// Step 12: Logs — query and stream execution logs
	// ---- Logs (initialized early for use in tasks group) ----
	logHandler := handlers.NewLogHandler(svc.LogService())
	logs := r.Group("/api/logs")
	{
		logs.GET("", logHandler.Query)
	}

	// Step 13: Tasks — workflow execution tracking (CRUD + lifecycle)
	// ---- Tasks ----
	taskHandler := handlers.NewTaskHandler(svc.TaskService())
	tasks := r.Group("/api/tasks")
	{
		tasks.GET("", taskHandler.List)
		tasks.GET("/:id", taskHandler.Get)
		tasks.POST("", taskHandler.Create)
		tasks.PUT("/:id/cancel", taskHandler.Cancel)
		tasks.PUT("/:id/status", taskHandler.UpdateStatus)
		tasks.DELETE("/:id", taskHandler.Delete)
		tasks.GET("/logs/:taskId", logHandler.FetchTaskLogs)
	}
	// Legacy task routes (redirected for backward compatibility)
	r.POST("/api/task/create", taskHandler.Create)
	r.GET("/api/task/:id/status", taskHandler.GetStatus)

	// Step 14: Plugins — discover, install, execute, status
	// ---- Plugins (V2) ----
	pluginHandler := handlers.NewPluginHandler(svc.PluginService())
	plugins := r.Group("/api/plugins")
	{
		plugins.GET("", pluginHandler.List)
		plugins.GET("/:name", pluginHandler.Get)
		plugins.GET("/nodes", pluginHandler.GetNodes)
		plugins.PUT("/:name/status", pluginHandler.UpdateStatus)
		plugins.POST("/install", pluginHandler.Install)
		plugins.DELETE("/:name", pluginHandler.Uninstall)
		plugins.GET("/:name/status", pluginHandler.InstallStatus)
		plugins.POST("/:name/execute", pluginHandler.Execute)
	}

	// Step 15: Agent — AI-powered chat, plan, generate workflow
	// ---- Agent ----
	agentHandler := handlers.NewAgentHandler(svc.AgentService())
	agent := r.Group("/api/agent")
	{
		agent.POST("/chat", agentHandler.Chat)
		agent.POST("/plan", agentHandler.PlanOnly)
		agent.POST("/generate-workflow", agentHandler.GenerateWorkflow)
		agent.POST("/generate-and-run", agentHandler.GenerateAndRunWorkflow)
	}

	// Step 16: MCP — Model Context Protocol server management
	// ---- MCP ----
	mcpHandler := handlers.NewMCPHandler(svc.MCPService())
	mcpGroup := r.Group("/api/mcp")
	{
		mcpGroup.GET("/tools", mcpHandler.ListTools)
		mcpGroup.GET("/servers", mcpHandler.ListServers)
		mcpGroup.GET("/status", mcpHandler.GetStatus)
		mcpGroup.GET("/config", mcpHandler.ExportConfig)
		mcpGroup.POST("/connect", mcpHandler.Connect)
		mcpGroup.POST("/disconnect", mcpHandler.Disconnect)
		mcpGroup.POST("/call", mcpHandler.Call)
		mcpGroup.POST("/servers", mcpHandler.AddServer)
		mcpGroup.DELETE("/servers/:name", mcpHandler.RemoveServer)
	}

	// Step 17: Environment — detect, check, repair, install deps
	// ---- Environment ----
	envHandler := handlers.NewEnvironmentHandler(svc.EnvironmentService())
	environment := r.Group("/api/environment")
	{
		environment.GET("/status", envHandler.GetStatus)
		environment.POST("/check", envHandler.Check)
		environment.GET("/repair-plan", envHandler.GetRepairPlan)
		environment.POST("/repair", envHandler.Repair)
		environment.POST("/install", envHandler.InstallDependency)
		environment.GET("/logs", envHandler.GetLogs)
		environment.DELETE("/logs", envHandler.ClearLogs)
	}

	// Step 18: Error Analysis — AI-powered error analysis and repair
	// ---- Error Analysis ----
	errorHandler := handlers.NewErrorAnalysisHandler(svc.LogService())
	errorGroup := r.Group("/api/error")
	{
		errorGroup.POST("/analyze", errorHandler.AnalyzeError)
		errorGroup.POST("/repair", errorHandler.RepairError)
		errorGroup.GET("/analysis/:taskId", errorHandler.GetErrorAnalysis)
		errorGroup.GET("/fix/:fixId/status", errorHandler.GetFixStatus)
	}

	// Step 19: Settings — application and engine configuration
	// ---- Settings ----
	settingsHandler := handlers.NewSettingsHandler()
	settings := r.Group("/api/settings")
	{
		settings.GET("", settingsHandler.GetSettings)
		settings.PUT("", settingsHandler.UpdateSettings)
		settings.GET("/engine", settingsHandler.GetEngineConfig)
		settings.PUT("/engine", settingsHandler.UpdateEngineConfig)
		settings.POST("/engine/test", settingsHandler.TestEngineConnection)
	}

	// Step 20: WebSocket — real-time bidirectional event stream
	// ---- WebSocket ----
	origins := resolveCORSOrigins()
	upgrader := newUpgrader(origins)

	// WebSocket handler: upgrade HTTP → WS, register client with Hub
	wsHandler := func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		if userID == nil {
			userID = "anonymous"
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := ws.NewClient(wsHub, conn, userID.(string), []string{"user:" + userID.(string)})
		wsHub.RegisterClient(client)

		go client.WritePump()
		go client.ReadPump()
	}

	r.GET("/api/ws", wsHandler)
	r.GET("/api/v1/ws", wsHandler)

	return r
}

// ============================================================================
// CORS Origin Resolution — desktop app + dev server origins
// ============================================================================

// resolveCORSOrigins returns allowed CORS origins based on environment.
// Supports: Tauri desktop, Vite dev server, and custom CORS_ALLOWED_ORIGINS env var.
func resolveCORSOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw != "" {
		var origins []string
		for _, o := range strings.Split(raw, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
		return origins
	}
	origins := []string{"tauri://localhost", "https://tauri.localhost", "capacitor://localhost"}
	env := os.Getenv("AISTUDIO_ENV")
	if env == "" {
		env = "development"
	}
	if env == "development" {
		origins = append(origins, "http://localhost:5173", "http://localhost:5174", "http://localhost:3000", "http://localhost:8080", "http://localhost")
	}
	return origins
}
