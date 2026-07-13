package cloud

type Service struct {
	Config       CloudConfig
	Sync         SyncService
	License      *LicenseManager
	PluginMarket MarketService
}

func NewService(config CloudConfig) *Service {
	return &Service{
		Config:       config,
		Sync:         NewSyncService(),
		License:      NewLicenseManager(),
		PluginMarket: NewMarketService(),
	}
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() error {
	return nil
}
