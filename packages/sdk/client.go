package sdk

import (
	"github.com/aistudio/packages/agent"
	"github.com/aistudio/packages/cloud"
	"github.com/aistudio/packages/compiler"
	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/plugin"
	"github.com/aistudio/packages/project"
	"github.com/aistudio/packages/runtime"
	"github.com/aistudio/packages/skill"
)

type ClientConfig struct {
	CloudConfig    cloud.CloudConfig
	ProjectsDir    string
	EventBus       *event.EventBus
	LLMProvider    agent.LLMProvider
}

func NewClient(config ClientConfig) (*Client, error) {
	c := &Client{
		eventBus: config.EventBus,
	}

	c.Cloud = cloud.NewService(config.CloudConfig)
	c.Compiler = compiler.NewCompiler(config.EventBus)

	if config.ProjectsDir != "" {
		c.Project = project.NewManager(config.ProjectsDir)
	}

	c.PluginRegistry = plugin.NewRegistry()
	c.Runtime = runtime.NewLocalExecutor()

	skillMgr := skill.NewSkillManager()
	c.Agent = agent.NewAgent(skillMgr)

	if config.LLMProvider != nil {
		c.Agent.WithLLM(config.LLMProvider)
	}
	if config.EventBus != nil {
		c.Agent.WithEventBus(config.EventBus)
	}

	return c, nil
}