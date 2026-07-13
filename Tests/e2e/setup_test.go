package e2e

import (
	"os"
	"sync"

	"github.com/aistudio/backend/internal/api"
	"github.com/aistudio/backend/internal/api/middleware"
	"github.com/aistudio/backend/internal/auth"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/database"
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
	"gorm.io/gorm"
)

var (
	testServices *service.Services
	testDB       *gorm.DB
	testRouter   *gin.Engine
	initOnce     sync.Once
)

func setupTestEnvironment() {
	initOnce.Do(func() {
		gin.SetMode(gin.TestMode)

		os.Setenv("DATABASE_TYPE", "sqlite")
		os.Setenv("DATABASE_URL", "file::memory:?cache=shared")
		os.Setenv("JWT_SECRET", "test-secret-do-not-use-in-production")

		_ = config.Reload()

		cfg := config.Get()
		cfg.Database.Type = "sqlite"
		cfg.Database.URL = "file::memory:?cache=shared"
		cfg.JWT.Secret = "test-secret-do-not-use-in-production"

		if err := database.Init(&database.Config{Type: "sqlite", URL: "file::memory:?cache=shared"}); err != nil {
			panic("database init failed: " + err.Error())
		}
		testDB = database.GetDB()

		bus := eventbus.New()

		logCenter := logcenter.New(1000)

		projectManager := project.NewManager(os.TempDir())

		runtimeExec := runtime.NewLocalExecutor()

		bundleManager := runtime.NewBundleManager(os.TempDir())

		skillManager := skill.NewManager()

		taskManager := task.NewManager(4)

		pluginManager := plugin.NewManager(os.TempDir())

		authManager := auth.NewManager(testDB, cfg.JWT.Secret, 0, 0)

		envManager := environment.NewManager()

		envIntegration := environment.NewEnvironmentIntegration(envManager, bundleManager, bus)

		engineClient := engine.NewClient(engine.DefaultConfig())

		container := service.NewContainer(service.ContainerParams{
			EventBus:         bus,
			LogCenter:        logCenter,
			ProjectManager:   projectManager,
			Compiler:         nil,
			Runtime:          runtimeExec,
			BundleManager:    bundleManager,
			Executor:         runtimeExec,
			SkillManager:     skillManager,
			PluginManager:    pluginManager,
			TaskManager:      taskManager,
			DiagnosticEngine: nil,
			AuthManager:      authManager,
			Authenticator:    authManager.Authenticator,
			DB:               testDB,
			Config:           cfg,
			EnvIntegration:   envIntegration,
			EngineClient:     engineClient,
		})

		testServices = service.NewServices(container)

		mwCfg := middleware.Config{
			JWTSecret:     cfg.JWT.Secret,
			AllowedOrigins: []string{"*"},
		}
		testRouter = api.SetupRouter(testServices, mwCfg)
	})
}
