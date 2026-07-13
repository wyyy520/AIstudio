package plugin

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// PluginRecord is the GORM database model for plugin persistence (V2).
type PluginRecord struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	PluginID  string    `gorm:"uniqueIndex;size:128;not null" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Version   string    `gorm:"size:32;not null" json:"version"`
	Author    string    `gorm:"size:128" json:"author"`
	Type      string    `gorm:"size:32" json:"category"`
	Status    string    `gorm:"size:32;not null;default:not_installed" json:"status"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
func (r *PluginRepository) Save(p *Plugin) error {
	record := r.pluginToRecord(p)
	result := r.db.Create(record)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[plugin-repo] saved plugin: %s", p.Name)
	return nil
}

// Update updates an existing plugin record.
func (r *PluginRepository) Update(p *Plugin) error {
	record := r.pluginToRecord(p)
	result := r.db.Model(&PluginRecord{}).Where("name = ?", p.Name).Updates(record)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[plugin-repo] updated plugin: %s", p.Name)
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
func (r *PluginRepository) pluginToRecord(p *Plugin) *PluginRecord {
	return &PluginRecord{
		PluginID: p.ID,
		Name:     p.Name,
		Version:  p.Version,
		Author:   p.Author,
		Type:     string(p.Type),
		Status:   string(p.Status),
		Enabled:  p.Enabled,
	}
}
