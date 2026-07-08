package api

import (
	"github.com/aistudio/backend/internal/api/handlers"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin router with all API routes.
// It takes the aggregated Services struct and middleware config, then wires everything together.
// Middleware is applied via middleware.Apply() for consistent configuration.
func SetupRouter(svc *service.Services, mwCfg middleware.Config) *gin.Engine {
	r := gin.New()

	// ---- Global middleware (unified registration) ----
	// Order: Recovery → Logger → CORS → RateLimit → Auth
	// Configure via middleware.Config or environment variables
	middleware.Apply(r, mwCfg)

	// ---- Health ----
	healthHandler := handlers.NewHealthHandler(svc)
	r.GET("/api/health", healthHandler.Check)

	// ---- Auth (login is public) ----
	authHandler := handlers.NewAuthHandler()
	r.POST("/api/auth/login", authHandler.Login)

	// ---- Users ----
	userHandler := handlers.NewUserHandler(svc.User)
	users := r.Group("/api/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.POST("", userHandler.Create)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
	}

	// ---- Projects ----
	projectHandler := handlers.NewProjectHandler(svc.Project)
	projects := r.Group("/api/projects")
	{
		projects.GET("", projectHandler.List)
		projects.GET("/:id", projectHandler.Get)
		projects.POST("", projectHandler.Create)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
	}

	// ---- Workflows ----
	workflowHandler := handlers.NewWorkflowHandler(svc.Workflow)
	workflowHandler.SetTaskService(svc.Task)
	workflows := r.Group("/api/workflows")
	{
		workflows.GET("", workflowHandler.List)
		workflows.GET("/:id", workflowHandler.Get)
		workflows.POST("", workflowHandler.Create)
		workflows.PUT("/:id", workflowHandler.Update)
		workflows.DELETE("/:id", workflowHandler.Delete)
		// Preserved: workflow run API
		workflows.POST("/:id/run", workflowHandler.Run)
		// Preserved: node list API
		workflows.GET("/nodes", workflowHandler.ListNodeTypes)
	}

	// ---- Tasks ----
	taskHandler := handlers.NewTaskHandler(svc.Task)
	// POST /api/task/create
	r.POST("/api/task/create", taskHandler.Create)
	// GET /api/task/:id/status
	r.GET("/api/task/:id/status", taskHandler.GetStatus)
	tasks := r.Group("/api/tasks")
	{
		tasks.GET("", taskHandler.List)
		tasks.GET("/:id", taskHandler.Get)
		tasks.POST("", taskHandler.Create)
		tasks.PUT("/:id/cancel", taskHandler.Cancel)
		tasks.PUT("/:id/status", taskHandler.UpdateStatus)
		tasks.DELETE("/:id", taskHandler.Delete)
	}

	// ---- Plugins ----
	pluginHandler := handlers.NewPluginHandler(svc.Plugin)
	// GET /api/plugins - list all plugins
	r.GET("/api/plugins", pluginHandler.GetPlugins)
	// POST /api/plugin/install - install a plugin
	r.POST("/api/plugin/install", pluginHandler.InstallPlugin)
	// POST /api/plugin/remove - remove a plugin
	r.POST("/api/plugin/remove", pluginHandler.RemovePlugin)
	// GET /api/plugin/:id - get plugin details
	r.GET("/api/plugin/:id", pluginHandler.GetPluginByID)
	plugins := r.Group("/api/plugins")
	{
		plugins.GET("", pluginHandler.List)
		plugins.GET("/:name", pluginHandler.Get)
		plugins.POST("/install", pluginHandler.Install)
		plugins.PUT("/:name/status", pluginHandler.UpdateStatus)
		plugins.DELETE("/:name", pluginHandler.Uninstall)
		plugins.POST("/:name/execute", pluginHandler.Execute)
	}

	// ---- Agent ----
	agentHandler := handlers.NewAgentHandler(svc.Agent)
	agent := r.Group("/api/agent")
	{
		agent.POST("/chat", agentHandler.Chat)
		agent.POST("/plan", agentHandler.PlanOnly)
	}

	// ---- MCP ----
	mcpHandler := handlers.NewMCPHandler(svc.MCP)
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

	// ---- Logs ----
	logHandler := handlers.NewLogHandler(svc.Log)
	logs := r.Group("/api/logs")
	{
		logs.GET("", logHandler.Query)
	}

	// ---- Environment ----
	envHandler := handlers.NewEnvironmentHandler(svc.Environment)
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

	// ---- WebSocket ----
	wsHandler := handlers.NewWebSocketHandler(svc.Task)
	r.GET("/api/ws", func(c *gin.Context) {
		wsHandler.HandleWebSocket(c.Writer, c.Request)
	})

	return r
}