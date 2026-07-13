package sdk

import (
	"github.com/aistudio/packages/agent"
	"github.com/aistudio/packages/cloud"
	"github.com/aistudio/packages/compiler"
	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/plugin"
	"github.com/aistudio/packages/project"
	"github.com/aistudio/packages/runtime"
)

type Client struct {
	Cloud          *cloud.Service
	Compiler       compiler.Compiler
	Project        *project.Manager
	PluginRegistry *plugin.Registry
	Runtime        runtime.CommandExecutor
	Agent          *agent.Agent

	eventBus *event.EventBus
}