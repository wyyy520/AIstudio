// AIStudio Backend Server
//
// AIStudio is a Visual AI Engineering Platform.
// It transforms Workflow DSL into real, runnable engineering projects.
//
// Architecture:
//   Workflow 鈫?Compiler 鈫?Generator 鈫?Project 鈫?Runtime 鈫?Log 鈫?Diagnostic
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/aistudio/backend/internal/api"
	auth "github.com/aistudio/backend/internal/auth"
	apiMiddleware "github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/api/handlers"
	ws "github.com/aistudio/backend/internal/api/ws"
	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/compiler/generators"
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
	"github.com/aistudio/backend/internal/workflow"
)

//go:embed web
var embeddedFS embed.FS

func main() {
	// ============================================================================
	// Initialize Configuration
	// ============================================================================
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg := config.Get()

	// Enforce non-default JWT secret before starting
	auth.MustNotUseDefaultSecret()

	// ============================================================================
	// Initialize Event Bus (Foundation for all module communication)
	// ===========================================================================
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

	// Log Center 鈥?unified log management
	logCenter := logcenter.NewPersistent(10000, "logs")

	// Project Manager 鈥?filesystem-based project management
	projectManager := project.NewManager("projects")

	// Skill Manager 鈥?workflow templates
	skillManager := skill.NewManager()
	loadBuiltinSkills(skillManager)

	// Compiler 鈥?workflow 鈫?project compilation
	compilerEngine := compiler.NewCompiler(bus)
	registerGenerators(compilerEngine)

	// Runtime 鈥?unified execution engine
	runtimeExecutor := runtime.NewExecutor(runtime.ExecutorLocal)
	runtimeEngine := runtime.NewLocalRuntime(runtimeExecutor)

	// Environment 鈥?environment manager
	envManager := environment.NewManager()
	bundleManager := runtime.NewBundleManager("project_bundles")
	envIntegration := environment.NewEnvironmentIntegration(envManager, bundleManager, bus)

	// Plugin Manager V2 鈥?pure declarative plugin discovery
	pluginManager := plugin.NewManager(cfg.Plugin.Directory)
	pluginManager.DiscoverPlugins()

	// Diagnostic Engine 鈥?AI-powered error analysis
	diagnosticEngine := diagnostic.NewEngine(db)

	// Task Manager 鈥?workflow execution tracking
	taskManager := task.NewManager(cfg.Task.NumWorkers)

	// Engine Client — bridge to Python AI Engine
	engineClient := engine.NewClient(&engine.Config{
		BaseURL:    cfg.Engine.URL,
		Timeout:    cfg.Engine.Timeout,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	// Inject EngineClient into workflow package for AI node executors
	workflow.SetEngineClient(engineClient)

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
	setupEventSubscriptions(bus, logCenter, diagnosticEngine)
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
	handler := buildHandler(cfg, services, authMgr.Authenticator, wsHub)

	// ============================================================================
	// Start Server
	// ============================================================================
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      handler,
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
// HTTP Handler
// ============================================================================

func buildHandler(cfg *config.Config, services *service.Container, authenticator *auth.Authenticator, wsHub *ws.Hub) http.Handler {
	svc := service.NewServices(services)

	mwCfg := apiMiddleware.DefaultConfig()
	mwCfg.JWTSecret = cfg.JWT.Secret
	mwCfg.Development = true

	// Gin router for API routes only
	ginRouter := api.SetupRouter(svc, mwCfg, wsHub)

	// Additional API routes
	healthHandler := handlers.NewHealthHandler(svc)
	ginRouter.GET("/api/v1/health", healthHandler.Check)

	// ============================================================================
	// Embedded frontend from the Go binary
	// ============================================================================
	subFS, err := fs.Sub(embeddedFS, "web")
	if err != nil {
		log.Fatalf("Failed to read embedded web fs: %v", err)
	}

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		data, err := fs.ReadFile(subFS, cleanPath)
		if err != nil {
			// SPA fallback
			data, err = fs.ReadFile(subFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}

		contentType := "application/octet-stream"
		switch path.Ext(cleanPath) {
		case ".js":
			contentType = "application/javascript"
		case ".css":
			contentType = "text/css"
		case ".png":
			contentType = "image/png"
		case ".svg":
			contentType = "image/svg+xml"
		case ".ico":
			contentType = "image/x-icon"
		case ".woff2":
			contentType = "font/woff2"
		case ".html":
			contentType = "text/html; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") {
			ginRouter.ServeHTTP(w, r)
		} else {
			staticHandler.ServeHTTP(w, r)
		}
	})
}

// ============================================================================
// Generators
// ============================================================================

func registerGenerators(compilerEngine compiler.Compiler) {
	compilerEngine.RegisterGenerator(generators.NewPythonAdapter())
	log.Println("[main] registered Python generator (template-based)")

	compilerEngine.RegisterGenerator(generators.NewMATLABAdapter())
	log.Println("[main] registered MATLAB generator")
	compilerEngine.RegisterGenerator(generators.NewROS2Adapter())
	log.Println("[main] registered ROS2 generator")
	compilerEngine.RegisterGenerator(generators.NewDockerAdapter())
	log.Println("[main] registered Docker generator")
	compilerEngine.RegisterGenerator(generators.NewSTM32Adapter())
	log.Println("[main] registered STM32 generator")
	compilerEngine.RegisterGenerator(generators.NewCPPAdapter())
	log.Println("[main] registered C++ generator")
	compilerEngine.RegisterGenerator(generators.NewUnityAdapter())
	log.Println("[main] registered Unity generator")
	compilerEngine.RegisterGenerator(generators.NewJavaAdapter())
	log.Println("[main] registered Java generator")
}

func loadBuiltinSkills(skillManager *skill.Manager) {
	err := skillManager.LoadFromDir("internal/skill/templates")
	if err != nil {
		log.Printf("[main] warning: could not load skill templates: %v", err)
	}
	log.Printf("[main] loaded %d built-in skills", skillManager.Count())
}

// ============================================================================
// Event Subscriptions
// ============================================================================

func setupEventSubscriptions(bus *eventbus.EventBus, logCenter logcenter.Logger, diagnosticEngine *diagnostic.Engine) {
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

			// Auto-analyze error-level logs with Diagnose Center
			if level == logcenter.LevelError && diagnosticEngine != nil && data.TaskID != "" {
				entry := &logcenter.Entry{
					Level:      level,
					Source:     data.Source,
					Message:    data.Message,
					TaskID:     data.TaskID,
					WorkflowID: data.WorkflowID,
					NodeID:     data.NodeID,
					Raw:        data.Raw,
				}
				result, err := diagnosticEngine.Analyze(context.Background(), entry, nil)
				if err == nil && result != nil {
					logCenter.Warn("diagnose", fmt.Sprintf("Diagnosis: %s - Suggestion: %s", result.Summary, result.RootCause))
				}
			}
		}
	})

	// Log compiler events
	bus.Subscribe(eventbus.TopicCompileStarted, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			logCenter.Info("compiler", fmt.Sprintf("Compilation started: %s -> %s", data.WorkflowID, data.Target))
		}
	})

	bus.Subscribe(eventbus.TopicCompileCompleted, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			logCenter.Info("compiler", fmt.Sprintf("Compilation completed: %s (duration=%s)", data.WorkflowID, data.Duration))
		}
	})

	bus.Subscribe(eventbus.TopicCompileFailed, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			logCenter.Error("compiler", fmt.Sprintf("Compilation failed: %s - %s", data.WorkflowID, data.Error))
		}
	})

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

	// Compile progress events → WebSocket
	eventBus.Subscribe(eventbus.TopicCompileStarted, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeCompileProgress, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("workflow:"+data.WorkflowID, dataBytes)
		}
	})

	eventBus.Subscribe(eventbus.TopicCompileProgress, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeCompileProgress, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("workflow:"+data.WorkflowID, dataBytes)
		}
	})

	eventBus.Subscribe(eventbus.TopicCompileCompleted, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeCompileProgress, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("workflow:"+data.WorkflowID, dataBytes)
		}
	})

	eventBus.Subscribe(eventbus.TopicCompileFailed, func(e eventbus.Event) {
		if data, ok := e.Data.(eventbus.CompileEventData); ok {
			msg := ws.NewMessage(ws.MsgTypeCompileProgress, data)
			dataBytes, _ := msg.ToJSON()
			hub.BroadcastToRoom("workflow:"+data.WorkflowID, dataBytes)
		}
	})
}

