package cloud

import "time"

type CloudConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}

type SyncStatus struct {
	Syncing  bool      `json:"syncing"`
	Idle     bool      `json:"idle"`
	Error    string    `json:"error,omitempty"`
	LastSync time.Time `json:"last_sync"`
}

type LicenseInfo struct {
	Licensed  bool      `json:"licensed"`
	ExpiresAt time.Time `json:"expires_at"`
	Features  []string  `json:"features"`
	MaxUsers  int       `json:"max_users"`
}

type RemoteCache struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
	Region   string `json:"region"`
}

type PluginInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Version     string  `json:"version"`
	Author      string  `json:"author"`
	Description string  `json:"description"`
	Downloads   int     `json:"downloads"`
	Rating      float64 `json:"rating"`
}
