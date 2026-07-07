package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aistudio/backend/internal/api"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/database"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/service"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
)

func main() {
	log.Println("=== AIStudio Backend ===")

	// ---- 0. Load Configuration ----
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg := config.Get()
	log.Printf("[config] server=%s database=%s log_level=%s",
		cfg.Server.Addr(), cfg.Database.Type, cfg.Log.Level)

	// ---- 1. Initialize Database ----
	dbCfg := &database.Config{
		Type: cfg.Database.Type,
		URL:  cfg.Database.URL,
	}
	if err := database.Init(dbCfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	db := database.GetDB()

	// ---- 2. Initialize Task Manager ----
	taskMgr := task.NewManager(cfg.Task.NumWorkers)
	taskMgr.Start()
	defer taskMgr.Stop()

	// ---- 3. Initialize Plugin Manager ----
	pluginMgr := plugin.NewManager(cfg.Plugin.Directory)
	if err := pluginMgr.DiscoverPlugins(); err != nil {
		log.Printf("Warning: Plugin discovery failed: %v", err)
	}

	// ---- 4. Register Workflow Engine Node Types ----
	workflow.RegisterDefaultNodes()
	engine := workflow.NewDefaultEngine()

	// Register a workflow task handler
	workflowTaskHandler := workflow.NewTaskHandler(engine)
	taskMgr.RegisterHandler("workflow", workflowTaskHandler)
	taskMgr.RegisterHandler("agent", workflowTaskHandler)

	// ---- 4.5 Initialize Environment Manager ----
	envMgr := environment.NewManager()

	// ---- 5. Initialize Services ----
	svc := service.NewServices(db, taskMgr, pluginMgr, engine, envMgr)

	// ---- 6. Setup HTTP Server ----
	mwCfg := middleware.DefaultConfig()
	mwCfg.JWTSecret = cfg.JWT.Secret
	mwCfg.Development = os.Getenv("AISTUDIO_ENV") == "development" || cfg.Log.Level == "debug"
	router := api.SetupRouter(svc, mwCfg)

	addr := cfg.Server.Addr()
	log.Printf("AIStudio Backend starting on %s", addr)
	log.Printf("API Documentation: http://localhost%s/api/health", addr)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
	}()

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}