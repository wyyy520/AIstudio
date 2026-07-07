package service

import (
	"log"

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
	Log         *LogService
	User        *UserService
	Environment *EnvironmentService
}

// NewServices creates all services with their dependencies.
func NewServices(db *gorm.DB, taskMgr *task.Manager, pluginMgr *plugin.Manager, engine *workflow.Engine, envMgr *environment.Manager) *Services {
	log.Println("[service] initializing all services...")

	services := &Services{
		Project:     NewProjectService(db),
		Workflow:    NewWorkflowService(db, engine),
		Task:        NewTaskService(db, taskMgr),
		Plugin:      NewPluginService(db, pluginMgr),
		Agent:       NewAgentService(taskMgr),
		Log:         NewLogService(),
		User:        NewUserService(db),
		Environment: NewEnvironmentService(envMgr),
	}

	log.Println("[service] all services initialized")
	return services
}