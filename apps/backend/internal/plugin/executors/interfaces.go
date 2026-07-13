package executors

import (
	"context"

	"github.com/aistudio/backend/internal/plugin"
)

type PluginExecutor interface {
	Execute(ctx context.Context, p *plugin.Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)
	Language() string
}
