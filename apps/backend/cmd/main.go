// AIStudio Backend Server
//
// AIStudio is a Visual AI Engineering Platform.
// It transforms Workflow DSL into real, runnable engineering projects.
//
// Architecture:
//   Workflow → Compiler → Generator → Project → Runtime → Log → Diagnostic
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aistudio/backend/internal/api"
	apiMiddleware "github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/api/handlers"
	ws "github.com/aistudio/backend/internal/api/ws"
	"github.com/aistudio/backend/internal/auth"
	"github.com/aistudio/backend/internal/compiler"
	compilerPython "github.com/aistudio/backend/internal/compiler/generators/python"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/database"
	"github.com/aistudio/backend/internal/diagnostic"
	"github.com/aistudio/backend/internal/engine"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/eventbus"
	"github.com/aistudio/backend/internal/logcenter"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/project"
	"github.com/aistudio/backend/internal/runtime"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/skill"
	"github.com/aistudio/backend/internal/task"
	"github.com/gin-gonic/gin"
)

func main() {
	// ============================================================================
	// Initialize Configuration
	// ============================================================================
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg := config.Get()

	// ============================================================================
	// Initialize Event Bus (Foundation for all module communication)
	// ============================================================================
	bus := eventbus.New(
		eventbus.WithHistorySize(1000),
		eventbus.WithTrace(false),
	)
	defer bus.Close()

	// ============================================================================
	// Initialize Data Layer
	// ============================================================================
	if err := database.Init(&database.Config{
		Type: cfg.Database.Type,
		URL:  cfg.Database.URL,
	}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	db := database.GetDB()

	// ============================================================================
	// Initialize Core Modules
	// ============================================================================

	// Log Center — unified log management
	logCenter := logcenter.New(10000)

	// Project Manager — filesystem-based project management
	projectManager := project.NewManager("projects")

	// Skill Manager — workflow templates
	skillManager := skill.NewManager()
	loadBuiltinSkills(skillManager)

	// Compiler — workflow → project compilation
	compilerEngine := compiler.NewCompiler(bus)
	registerGenerators(compilerEngine)

	// Runtime — unified execution engine
	runtimeExecutor := runtime.NewExecutor(runtime.ExecutorLocal)
	runtimeEngine := runtime.NewLocalRuntime(runtimeExecutor)

	// Environment — environment manager
	envManager := environment.NewManager()
	bundleManager := runtime.NewBundleManager("project_bundles")
	envIntegration := environment.NewEnvironmentIntegration(envManager, bundleManager, bus)

	// Plugin Manager V2 — pure declarative plugin discovery
	pluginManager := plugin.NewManager(cfg.Plugin.Directory)
	pluginManager.DiscoverPlugins()

	// Diagnostic Engine — AI-powered error analysis
	diagnosticEngine := diagnostic.NewEngine(db)

	// Task Manager — workflow execution tracking
	taskManager := task.NewManager(cfg.Task.NumWorkers)

	// Engine Client — bridge to Python AI Engine
	engineClient := engine.NewClient(&engine.Config{
		BaseURL:    cfg.Engine.URL,
		Timeout:    cfg.Engine.Timeout,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	// Auth Manager — authentication and authorization
	authMgr := auth.NewManager(db, cfg.JWT.Secret, 24*time.Hour, 7*24*time.Hour)

	// ============================================================================
	// Initialize WebSocket Hub
	// ============================================================================
	wsHub := ws.NewHub()
	go wsHub.Run()

	// ============================================================================
	// Setup Event Bus Subscriptions
	// ============================================================================
	setupEventSubscriptions(bus, logCenter)
	setupWSSubscriptions(wsHub, bus)

	// ============================================================================
	// Initialize Service Layer
	// ============================================================================
	services := service.NewContainer(service.ContainerParams{
		EventBus:         bus,
		LogCenter:        logCenter,
		ProjectManager:   projectManager,
		Compiler:         compilerEngine,
		Runtime:          runtimeEngine,
		BundleManager:    bundleManager,
		Executor:         runtimeExecutor,
		SkillManager:     skillManager,
		PluginManager:    pluginManager,
		TaskManager:      taskManager,
		DiagnosticEngine: diagnosticEngine,
		AuthManager:      authMgr.UserManager(),
		Authenticator:    authMgr.Authenticator,
		DB:               db,
		Config:           cfg,
		EnvIntegration:   envIntegration,
		EngineClient:     engineClient,
	})

	// ============================================================================
	// Initialize HTTP API
	// ============================================================================
	router := setupRouter(cfg, services, authMgr.Authenticator, wsHub)

	// ============================================================================
	// Start Server
	// ============================================================================
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		bus.Publish(eventbus.TopicSystemShutdown, eventbus.SystemEventData{
			Component: "server",
			Action:    "shutdown",
		})

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("AIStudio server starting on %s", cfg.Server.Addr())
	log.Printf("  Projects directory: projects")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server stopped")
}

// ============================================================================
// Setup Functions
// ============================================================================

func setupRouter(cfg *config.Config, services *service.Container, authenticator *auth.Authenticator, wsHub *ws.Hub) *gin.Engine {
	svc := service.NewServices(services)

	mwCfg := apiMiddleware.DefaultConfig()
	mwCfg.JWTSecret = cfg.JWT.Secret
	mwCfg.Development = true

	router := api.SetupRouter(svc, mwCfg, wsHub)

	// v1 health check
	healthHandler := handlers.NewHealthHandler(svc)
	router.GET("/api/v1/health", healthHandler.Check)

	return router
}

func registerGenerators(compilerEngine compiler.Compiler) {
	// Register Python Generator
	compilerEngine.RegisterGenerator(compilerPython.NewGenerator())
	log.Println("[main] registered Python generator")

	// TODO: Register more generators as they are implemented
	// compilerEngine.RegisterGenerator(matlab.NewGenerator())
	// compilerEngine.RegisterGenerator(ros2.NewGenerator())
	// compilerEngine.RegisterGenerator(docker.NewGenerator())
	// compilerEngine.RegisterGenerator(stm32.NewGenerator())
}

func loadBuiltinSkills(skillManager *skill.Manager) {
	// Load built-in skill templates from embedded files
	// For now, these are loaded from the templates directory
	err := skillManager.LoadFromDir("internal/skill/templates")
	if err != nil {
		log.Printf("[main] warning: could not load skill templates: %v", err)
	}

	log.Printf("[main] loaded %d built-in skills", skillManager.Count())
}

func setupEventSubscriptions(bus *eventbus.EventBus, logCenter *logcenter.LogCenter) {
	// Log all runtime events
	bus.Subscribe(eventbus.TopicRuntimeLog, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.LogEventData); ok {
			level := logcenter.LevelInfo
			switch data.Level {
			case "error", "ERROR":
				level = logcenter.LevelError
			case "warn", "WARN":
				level = logcenter.LevelWarn
			case "debug", "DEBUG":
				level = logcenter.LevelDebug
			}
			logCenter.LogEntry(logcenter.Entry{
				Level:      level,
				Source:     data.Source,
				Message:    data.Message,
				TaskID:     data.TaskID,
				WorkflowID: data.WorkflowID,
				NodeID:     data.NodeID,
				Raw:        data.Raw,
			})
		}
	})

	// Log system events
	bus.Subscribe(eventbus.TopicSystemStartup, func(e eventbus.Event) {
		logCenter.Info("system", "System started")
	})

	bus.Subscribe(eventbus.TopicSystemShutdown, func(e eventbus.Event) {
		logCenter.Info("system", "System shutting down")
	})
}

func setupWSSubscriptions(hub *ws.Hub, eventBus *eventbus.EventBus) {
	taskTopics := []eventbus.Topic{
		eventbus.TopicTaskCreated,
		eventbus.TopicTaskStarted,
		eventbus.TopicTaskCompleted,
		eventbus.TopicTaskFailed,
		eventbus.TopicTaskCancelled,
	}

	for _, topic := range taskTopics {
		t := topic
		eventBus.Subscribe(t, func(e eventbus.Event) {
			if data, ok := e.Data.(eventbus.TaskEventData); ok {
				msg := ws.NewMessage(ws.MsgTypeTaskStatus, data)
				dataBytes, _ := msg.ToJSON()
				hub.BroadcastToRoom("task:"+data.TaskID, dataBytes)
			}
		})
	}

	eventBus.Subscribe(eventbus.TopicTaskProgress, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.TaskEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeTaskStatus, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("task:"+data.TaskID, dataBytes)
		}
	})

	eventBus.Subscribe(eventbus.TopicRuntimeLog, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.RuntimeEventData); ok && data.TaskID != "" {
			msg := ws.NewMessage(ws.MsgTypeNodeLog, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("task:"+data.TaskID, dataBytes)
		}
		if data, ok := e.Data.(eventbus.LogEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeNodeLog, data)
			dataBytes, _ := msg.ToJSON()
			if data.TaskID != "" {
				hub.BroadcastToRoom("task:"+data.TaskID, dataBytes)
			}
			if data.WorkflowID != "" {
				hub.BroadcastToRoom("workflow:"+data.WorkflowID, dataBytes)
			}
		}
	})

	eventBus.Subscribe(eventbus.TopicRuntimeProgress, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.RuntimeEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeWorkflowProgress, data)
			dataBytes, _ := msg.ToJSON()
			if data.TaskID != "" {
				hub.BroadcastToRoom("task:"+data.TaskID, dataBytes)
			}
		}
	})
}