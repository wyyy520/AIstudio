package service

import (
	"log"
	"strconv"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/auth"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

type Services struct {
	Project     *ProjectService
	Workflow    *WorkflowService
	Task        *TaskService
	Plugin      *PluginService
	Agent       *AgentService
	MCP         *MCPService
	Log         *LogService
	User        *UserService
	Environment *EnvironmentService
	Auth        *auth.Authenticator
}

func NewServices(db *gorm.DB, taskMgr *task.Manager, pluginMgr *plugin.Manager, engine *workflow.Engine, envMgr *environment.Manager, ag *agent.Agent) *Services {
	log.Println("[service] initializing all services...")

	cfg := config.Get()
	mcpService := NewMCPService(cfg.MCP)

	authSecret := cfg.JWT.Secret
	if authSecret == "" {
		authSecret = "aistudio-default-secret-change-in-production"
	}
	auth.SetJWTSecret(authSecret)

	userMgr := auth.NewUserManager(db)
	tokenMgr := auth.NewTokenManager(24*60*60*1e9, 7*24*60*60*1e9)
	sessionMgr := auth.NewSessionManager(db)
	permMgr := auth.NewPermissionManager(db)
	quotaMgr := auth.NewQuotaManager(db)
	apiKeyMgr := auth.NewAPIKeyManager(db, authSecret)

	authenticator := auth.NewAuthenticator(
		userMgr, tokenMgr, sessionMgr, permMgr, quotaMgr, apiKeyMgr,
	)

	services := &Services{
		Project:     NewProjectService(db),
		Workflow:    NewWorkflowService(db, engine),
		Task:        NewTaskService(db, taskMgr),
		Plugin:      NewPluginService(db, pluginMgr),
		Agent:       NewAgentService(ag, pluginMgr, envMgr, engine, taskMgr, mcpService),
		MCP:         mcpService,
		Log:         NewLogService(),
		User:        NewUserService(userMgr),
		Environment: NewEnvironmentService(envMgr),
		Auth:        authenticator,
	}

	log.Println("[service] all services initialized")
	return services
}

func parseID(id string) (uint, error) {
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}
