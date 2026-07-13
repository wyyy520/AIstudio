package sdk

import (
	"github.com/aistudio/packages/plugin"
)

type PluginSummary = plugin.PluginSummary
type PluginNode = plugin.PluginNode

func ListPlugins() []PluginSummary {
	reg := plugin.NewRegistry()
	return reg.ListSummaries()
}

func GetNodeTypes(target string) []PluginNode {
	reg := plugin.NewRegistry()
	return reg.GetNodesByTarget(target)
}