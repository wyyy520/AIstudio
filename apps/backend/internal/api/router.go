package api

import (
	"net/http"

	"github.com/aistudio/backend/internal/api/handlers"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/api/ws"
	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetupRouter(svc *service.Services, mwCfg middleware.Config, wsHub *ws.Hub) *gin.Engine {
	r := gin.New()

	middleware.Apply(r, mwCfg)

	// ---- Health ----
	healthHandler := handlers.NewHealthHandler(svc)
	r.GET("/api/health", healthHandler.Check)

	// ---- Auth (public) ----
	authHandler := handlers.NewAuthHandler(svc.Authenticator)
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/refresh", authHandler.RefreshToken)

	// ---- Auth (authenticated) ----
	authGroup := r.Group("/api/auth")
	authGroup.POST("/logout", authHandler.Logout)

	// ---- User Profile ----
	profileHandler := handlers.NewProfileHandler(svc.Authenticator)
	profile := r.Group("/api/user")
	{
		profile.GET("/profile", profileHandler.GetProfile)
		profile.PUT("/profile", profileHandler.UpdateProfile)
		profile.GET("/sessions", profileHandler.GetSessions)
	}

	// ---- API Keys ----
	apiKeyHandler := handlers.NewAPIKeyHandler(svc.Authenticator)
	apiKeys := r.Group("/api/user/apikeys")
	{
		apiKeys.GET("", apiKeyHandler.List)
		apiKeys.POST("", apiKeyHandler.Create)
		apiKeys.DELETE("/:id", apiKeyHandler.Delete)
	}
	r.GET("/api/providers", apiKeyHandler.GetProviders)

	// ---- Quota ----
	quotaHandler := handlers.NewQuotaHandler(svc.Authenticator)
	quota := r.Group("/api/user/quota")
	{
		quota.GET("", quotaHandler.GetQuotas)
		quota.GET("/check", quotaHandler.CheckQuota)
	}

	// ---- Admin: Quota Management ----
	r.POST("/api/admin/quota", quotaHandler.UpdateQuota)

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

	// ---- Projects ----
	projectHandler := handlers.NewProjectHandler(svc.ProjectService())
	projects := r.Group("/api/projects")
	{
		projects.GET("", projectHandler.List)
		projects.GET("/:id", projectHandler.Get)
		projects.POST("", projectHandler.Create)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
	}

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

	// ---- Logs (initialized early for use in tasks group) ----
	logHandler := handlers.NewLogHandler(svc.LogService())
	logs := r.Group("/api/logs")
	{
		logs.GET("", logHandler.Query)
	}

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

	// ---- Agent ----
	agentHandler := handlers.NewAgentHandler(svc.AgentService())
	agent := r.Group("/api/agent")
	{
		agent.POST("/chat", agentHandler.Chat)
		agent.POST("/plan", agentHandler.PlanOnly)
		agent.POST("/generate-workflow", agentHandler.GenerateWorkflow)
		agent.POST("/generate-and-run", agentHandler.GenerateAndRunWorkflow)
	}

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

	// ---- Error Analysis ----
	errorHandler := handlers.NewErrorAnalysisHandler(svc.LogService())
	errorGroup := r.Group("/api/error")
	{
		errorGroup.POST("/analyze", errorHandler.AnalyzeError)
		errorGroup.POST("/repair", errorHandler.RepairError)
		errorGroup.GET("/analysis/:taskId", errorHandler.GetErrorAnalysis)
		errorGroup.GET("/fix/:fixId/status", errorHandler.GetFixStatus)
	}

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

	// ---- WebSocket ----
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
