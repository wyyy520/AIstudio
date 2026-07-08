package service

import (
	"log"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

// Services aggregates all domain services for easy wiring.
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
}

// NewServices creates all services with their dependencies.
func NewServices(db *gorm.DB, taskMgr *task.Manager, pluginMgr *plugin.Manager, engine *workflow.Engine, envMgr *environment.Manager, ag *agent.Agent) *Services {
	log.Println("[service] initializing all services...")

	// Initialize MCP service
	cfg := config.Get()
	mcpService := NewMCPService(cfg.MCP)

	services := &Services{
		Project:     NewProjectService(db),
		Workflow:    NewWorkflowService(db, engine),
		Task:        NewTaskService(db, taskMgr),
		Plugin:      NewPluginService(db, pluginMgr),
		Agent:       NewAgentService(ag, pluginMgr, envMgr, engine, taskMgr, mcpService),
		MCP:         mcpService,
		Log:         NewLogService(),
		User:        NewUserService(db),
		Environment: NewEnvironmentService(envMgr),
	}

	log.Println("[service] all services initialized")
	return services
}