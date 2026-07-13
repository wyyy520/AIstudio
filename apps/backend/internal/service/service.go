// Package service provides the business logic facade for AIStudio.
//
// Service is the thin layer between API handlers and core modules.
// It coordinates module interactions but contains no business logic itself.
package service

import (
	"github.com/aistudio/backend/internal/auth"
	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/compiler/generators"
	"github.com/aistudio/backend/internal/config"
	"github.com/aistudio/backend/internal/diagnostic"
	"github.com/aistudio/backend/internal/engine"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/eventbus"
	"github.com/aistudio/backend/internal/logcenter"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/project"
	"github.com/aistudio/backend/internal/runtime"
	"github.com/aistudio/backend/internal/skill"
	"github.com/aistudio/backend/internal/task"
	"gorm.io/gorm"
)

// ============================================================================
// Service Container
// ============================================================================

// Container holds all service instances and provides access to core modules.
// It is the single point of dependency injection for the API layer.
type Container struct {
	EventBus           *eventbus.EventBus
	LogCenter          *logcenter.LogCenter
	ProjectManager     *project.Manager
	Compiler           compiler.Compiler
	Runtime            runtime.Runtime
	BundleManager      runtime.BundleManager
	Executor           runtime.CommandExecutor
	SkillManager       *skill.Manager
	PluginManager      *plugin.Manager
	TaskManager        *task.Manager
	DiagnosticEngine   diagnostic.Diagnostic
	AuthManager        *auth.UserManager
	Authenticator      *auth.Authenticator
	DB                 *gorm.DB
	Config             *config.Config
	EnvIntegration     *environment.EnvironmentIntegration
	EngineClient       engine.EngineClient

	// lazy-loaded singletons
	workflowService *WorkflowService
}

// ContainerParams defines the dependencies for creating a Container.
type ContainerParams struct {
	EventBus         *eventbus.EventBus
	LogCenter        *logcenter.LogCenter
	ProjectManager   *project.Manager
	Compiler         compiler.Compiler
	Runtime          runtime.Runtime
	BundleManager    runtime.BundleManager
	Executor         runtime.CommandExecutor
	SkillManager     *skill.Manager
	PluginManager    *plugin.Manager
	TaskManager      *task.Manager
	DiagnosticEngine diagnostic.Diagnostic
	AuthManager      *auth.UserManager
	Authenticator    *auth.Authenticator
	DB               *gorm.DB
	Config         *config.Config
	EnvIntegration *environment.EnvironmentIntegration
	EngineClient   engine.EngineClient
}

// NewContainer creates a new Service Container.
func NewContainer(params ContainerParams) *Container {
	// Register built-in generators with the compiler
	if params.Compiler != nil {
		params.Compiler.RegisterGenerator(generators.NewPythonAdapter())
		params.Compiler.RegisterGenerator(generators.NewMATLABAdapter())
		params.Compiler.RegisterGenerator(generators.NewROS2Adapter())
		params.Compiler.RegisterGenerator(generators.NewDockerAdapter())
		params.Compiler.RegisterGenerator(generators.NewSTM32Adapter())
		params.Compiler.RegisterGenerator(generators.NewCPPAdapter())
		params.Compiler.RegisterGenerator(generators.NewUnityAdapter())
		params.Compiler.RegisterGenerator(generators.NewJavaAdapter())
	}

	// Build EnvironmentIntegration if not provided
	envIntegration := params.EnvIntegration
	if envIntegration == nil && params.BundleManager != nil {
		envManager := environment.NewManager()
		envIntegration = environment.NewEnvironmentIntegration(envManager, params.BundleManager, params.EventBus)
	}

	return &Container{
		EventBus:         params.EventBus,
		LogCenter:        params.LogCenter,
		ProjectManager:   params.ProjectManager,
		Compiler:         params.Compiler,
		Runtime:          params.Runtime,
		BundleManager:    params.BundleManager,
		Executor:         params.Executor,
		SkillManager:     params.SkillManager,
		PluginManager:    params.PluginManager,
		TaskManager:      params.TaskManager,
		DiagnosticEngine: params.DiagnosticEngine,
		AuthManager:      params.AuthManager,
		Authenticator:   params.Authenticator,
		DB:               params.DB,
		Config:           params.Config,
		EnvIntegration:   envIntegration,
		EngineClient:     params.EngineClient,
	}
}

// ============================================================================
// Service Interface
// ============================================================================

// Services provides access to all business services.
// This is a convenience wrapper for the Container.
type Services struct {
	*Container
}

// NewServices creates a new Services wrapper.
func NewServices(container *Container) *Services {
	return &Services{Container: container}
}

// ============================================================================
// Keep existing service implementations
// ============================================================================

// ProjectService provides project-related business logic.
func (s *Services) ProjectService() *ProjectService {
	return NewProjectService(s.DB, s.ProjectManager)
}

// WorkflowService provides workflow-related business logic.
// Uses lazy-loaded singleton to ensure plugin/node registry connections are preserved.
func (s *Services) WorkflowService() *WorkflowService {
	if s.workflowService == nil {
		s.workflowService = NewWorkflowService(s.DB, s.ProjectManager)
		// Connect plugin manager to workflow engine's node registry.
		// Must happen after WorkflowService is created so the engine exists.
		if s.PluginManager != nil && s.workflowService.Engine() != nil {
			s.PluginManager.SetWorkflowRegistry(s.workflowService.Engine().Registry())
		}
	}
	return s.workflowService
}

// AgentService provides agent-related business logic.
func (s *Services) AgentService() *AgentService {
	return NewAgentService(s.SkillManager)
}

// LogService provides log-related business logic.
func (s *Services) LogService() *LogService {
	return NewLogService(s.LogCenter)
}

// PluginService provides plugin-related business logic.
func (s *Services) PluginService() *PluginService {
	return NewPluginService(s.DB, s.PluginManager)
}

// TaskService provides task-related business logic.
func (s *Services) TaskService() *TaskService {
	return NewTaskService(s.DB, s.TaskManager)
}

// UserService provides user-related business logic.
func (s *Services) UserService() *UserService {
	return NewUserService(s.AuthManager)
}

// EnvironmentService provides environment-related business logic.
func (s *Services) EnvironmentService() *EnvironmentService {
	return NewEnvironmentService(s.Container)
}

// EnvironmentIntegration returns the integration bridge.
func (s *Services) EnvironmentIntegration() *environment.EnvironmentIntegration {
	return s.EnvIntegration
}

// MCPService provides MCP-related business logic.
func (s *Services) MCPService() *MCPService {
	return NewMCPService(s.Config.MCP)
}

// RuntimeService provides runtime execution and bundle management.
func (s *Services) RuntimeService() *RuntimeService {
	return NewRuntimeService(s.Runtime, s.BundleManager, s.Executor, s.Compiler, s.ProjectManager)
}

// BundleService provides bundle management (install, list, uninstall, clean).
func (s *Services) BundleService() *BundleService {
	return NewBundleService(s.BundleManager)
}

// EngineService returns the engine client for AI Engine (Python) communication.
func (s *Services) EngineService() engine.EngineClient {
	return s.EngineClient
}