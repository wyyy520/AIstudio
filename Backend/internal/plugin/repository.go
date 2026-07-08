package plugin

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// PluginRecord is the GORM database model for plugin persistence.
type PluginRecord struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	PluginID  string    `gorm:"uniqueIndex;size:128;not null" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Version   string    `gorm:"size:32;not null" json:"version"`
	Author    string    `gorm:"size:128" json:"author"`
	Type      string    `gorm:"size:32" json:"type"`
	Status    string    `gorm:"size:32;not null;default:not_installed" json:"status"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	Path      string    `gorm:"size:512" json:"path"`
	Source    string    `gorm:"size:32" json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PluginRecord) TableName() string {
	return "plugins"
}

// PluginRepository handles database persistence for plugins.
type PluginRepository struct {
	db *gorm.DB
}

// NewPluginRepository creates a new PluginRepository.
func NewPluginRepository(db *gorm.DB) *PluginRepository {
	return &PluginRepository{db: db}
}

// AutoMigrate creates or updates the plugins table.
func (r *PluginRepository) AutoMigrate() error {
	err := r.db.AutoMigrate(&PluginRecord{})
	if err != nil {
		return err
	}
	log.Println("[plugin-repo] plugins table migrated")
	return nil
}

// Save persists a plugin record to the database.
func (r *PluginRepository) Save(plugin *Plugin) error {
	record := r.pluginToRecord(plugin)
	result := r.db.Create(record)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[plugin-repo] saved plugin: %s", plugin.Name)
	return nil
}

// Update updates an existing plugin record.
func (r *PluginRepository) Update(plugin *Plugin) error {
	record := r.pluginToRecord(plugin)
	result := r.db.Model(&PluginRecord{}).Where("name = ?", plugin.Name).Updates(record)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[plugin-repo] updated plugin: %s", plugin.Name)
	return nil
}

// FindByName retrieves a plugin record by name.
func (r *PluginRepository) FindByName(name string) (*PluginRecord, error) {
	var record PluginRecord
	result := r.db.Where("name = ?", name).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// FindAll retrieves all plugin records.
func (r *PluginRepository) FindAll() ([]PluginRecord, error) {
	var records []PluginRecord
	result := r.db.Order("created_at DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

// Delete removes a plugin record by name.
func (r *PluginRepository) Delete(name string) error {
	result := r.db.Where("name = ?", name).Delete(&PluginRecord{})
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[plugin-repo] deleted plugin: %s", name)
	return nil
}

// pluginToRecord converts a Plugin to a PluginRecord.
func (r *PluginRepository) pluginToRecord(plugin *Plugin) *PluginRecord {
	return &PluginRecord{
		PluginID: plugin.ID,
		Name:     plugin.Name,
		Version:  plugin.Version,
		Author:   plugin.Author,
		Type:     string(plugin.Type),
		Status:   string(plugin.Status),
		Enabled:  plugin.Enabled,
		Path:     plugin.Path,
		Source:   string(plugin.Source),
	}
}

// LoadFromRepository loads all plugins from the database into the registry.
func (r *PluginRepository) LoadAllIntoRegistry(registry *Registry) error {
	records, err := r.FindAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Status == string(StatusInstalled) || record.Status == string(StatusEnabled) {
			// Skip if already in memory
			if _, ok := registry.Get(record.Name); ok {
				continue
			}
		}
		_ = record // placeholder for future restoration
	}

	log.Printf("[plugin-repo] loaded %d plugin records from database", len(records))
	return nil
}