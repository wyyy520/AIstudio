package cloud

import "fmt"

type MarketService interface {
	Search(query string) ([]PluginInfo, error)
	Install(pluginID string) error
}

type marketService struct {
	plugins map[string]PluginInfo
}

func NewMarketService() MarketService {
	return &marketService{
		plugins: make(map[string]PluginInfo),
	}
}

func (m *marketService) Search(query string) ([]PluginInfo, error) {
	var results []PluginInfo
	for _, p := range m.plugins {
		if query == "" {
			results = append(results, p)
		}
	}
	return results, nil
}

func (m *marketService) Install(pluginID string) error {
	_, ok := m.plugins[pluginID]
	if !ok {
		return fmt.Errorf("plugin %s not found in marketplace", pluginID)
	}
	return nil
}
