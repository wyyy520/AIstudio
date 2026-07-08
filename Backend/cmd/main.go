package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/api"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/database"
	aiengine "github.com/aistudio/backend/internal/engine"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/launcher"
	"github.com/aistudio/backend/internal/mcp"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
)

func main() {
	log.Println("=== AIStudio Backend ===")

	// ---- 0. Create Launcher ----
	lm := launcher.NewLauncher()

	// ---- 1. Load Configuration ----
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg := config.Get()
	log.Printf("[config] server=%s database=%s log_level=%s",
		cfg.Server.Addr(), cfg.Database.Type, cfg.Log.Level)

	// ---- 2. Initialize Database ----
	dbCfg := &database.Config{
		Type: cfg.Database.Type,
		URL:  cfg.Database.URL,
	}
	if err := database.Init(dbCfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	db := database.GetDB()

	// ---- 3. Initialize Task Manager ----
	taskMgr := task.NewManager(cfg.Task.NumWorkers)

	taskRepo := task.NewTaskRepository(db)
	if err := taskRepo.AutoMigrate(); err != nil {
		log.Printf("Warning: Task repository migration failed: %v", err)
	}
	taskMgr.SetRepository(taskRepo)

	taskMgr.Start()
	defer taskMgr.Stop()

	lm.Register("task-manager", 10, func(ctx context.Context) error {
		log.Println("[launcher] task manager initialized")
		return nil
	}, func(ctx context.Context) error {
		taskMgr.Stop()
		return nil
	})

	// ---- 4. Initialize Plugin Manager ----
	pluginMgr := plugin.NewManager(cfg.Plugin.Directory)

	pluginRepo := plugin.NewPluginRepository(db)
	if err := pluginRepo.AutoMigrate(); err != nil {
		log.Printf("Warning: Plugin repository migration failed: %v", err)
	}
	pluginMgr.SetRepository(pluginRepo)

	lm.Register("plugin-manager", 20, func(ctx context.Context) error {
		log.Println("[launcher] initializing plugin manager...")
		if err := pluginMgr.DiscoverPlugins(); err != nil {
			log.Printf("Warning: Plugin discovery failed: %v", err)
		}
		return nil
	}, nil)

	// ---- 5. Initialize Workflow Engine ----
	engine := workflow.NewDefaultEngine()
	workflow.RegisterDefaultNodes()
	pluginMgr.GetRegistry().SetWorkflowRegistry(engine.Registry())

	if err := pluginMgr.DiscoverPlugins(); err != nil {
		log.Printf("Warning: Plugin discovery failed: %v", err)
	}

	workflowTaskHandler := workflow.NewTaskHandler(engine)
	workflowTaskHandler.SetTaskManager(taskMgr)
	taskMgr.RegisterHandler("workflow", workflowTaskHandler)
	taskMgr.RegisterHandler("agent", workflowTaskHandler)

	lm.Register("workflow-engine", 30, func(ctx context.Context) error {
		log.Println("[launcher] workflow engine initialized")
		return nil
	}, nil)

	// ---- 6. Initialize Environment Manager ----
	envMgr := environment.NewManager()
	workflowTaskHandler.SetEnvironmentChecker(envMgr)

	lm.Register("environment-manager", 40, func(ctx context.Context) error {
		log.Println("[launcher] environment manager initialized")
		return nil
	}, nil)

	// ---- 7. Initialize Python AI Engine ----
	engineRunner := aiengine.NewPythonRunner(cfg.Engine.PythonPath, cfg.Engine.EngineDir)
	engineHandler := aiengine.NewTaskHandler(engineRunner)
	engineHandler.SetTaskManager(taskMgr)
	taskMgr.RegisterHandler("engine", engineHandler)
	taskMgr.RegisterHandler("yolo", engineHandler)

	pluginMgr.GetExecutor().SetEngineRunner(engineHandler)

	log.Printf("[engine] Python runner initialized: python=%s engine_dir=%s",
		cfg.Engine.PythonPath, cfg.Engine.EngineDir)

	lm.Register("python-engine", 50, func(ctx context.Context) error {
		log.Println("[launcher] Python engine initialized")
		return nil
	}, func(ctx context.Context) error {
		log.Println("[launcher] shutting down Python engine...")
		return nil
	})

	// ---- 8. Initialize Agent ----
	agentMemory, err := agent.NewMemory(db)
	if err != nil {
		log.Printf("Warning: Agent memory initialization failed: %v", err)
	}

	var llmProvider agent.LLMProvider
	if cfg.LLM.Provider != "" && cfg.LLM.Provider != "mock" {
		log.Printf("[agent] LLM provider configured: %s (model: %s)", cfg.LLM.Provider, cfg.LLM.Model)
	} else {
		log.Println("[agent] No LLM provider configured, using rule-based planning")
	}

	aiAgent := agent.NewAgent(llmProvider, agentMemory)
	log.Println("[agent] AI Agent initialized")

	lm.Register("agent", 60, func(ctx context.Context) error {
		log.Println("[launcher] AI Agent initialized")
		return nil
	}, nil)

	// ---- 9. Initialize Services ----
	svc := service.NewServices(db, taskMgr, pluginMgr, engine, envMgr, aiAgent)

	// Wire MCP Runtime into Workflow Engine
	mcpRuntimeAdapter := mcp.NewRuntimeAdapter(svc.MCP.Manager())
	engine.Registry().SetMCPRuntime(mcpRuntimeAdapter)
	log.Println("[mcp] MCP runtime wired into workflow engine")

	lm.Register("services", 70, func(ctx context.Context) error {
		log.Println("[launcher] services initialized")
		return nil
	}, nil)

	// ---- 10. Setup HTTP Server ----
	mwCfg := middleware.DefaultConfig()
	mwCfg.JWTSecret = cfg.JWT.Secret
	mwCfg.Development = os.Getenv("AISTUDIO_ENV") == "development" || cfg.Log.Level == "debug"
	router := api.SetupRouter(svc, mwCfg)

	lm.Register("http-server", 100, func(ctx context.Context) error {
		log.Println("[launcher] HTTP server starting...")
		return nil
	}, func(ctx context.Context) error {
		log.Println("[launcher] HTTP server stopped")
		return nil
	})

	addr := cfg.Server.Addr()
	log.Printf("AIStudio Backend starting on %s", addr)
	log.Printf("API Documentation: http://localhost%s/api/health", addr)

	// ---- Start all modules ----
	ctx := context.Background()
	if err := lm.Start(ctx); err != nil {
		log.Fatalf("Failed to start modules: %v", err)
	}

	// ---- Graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down...")
		lm.Stop(ctx)
		os.Exit(0)
	}()

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}