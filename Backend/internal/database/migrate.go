package database

import (
	"log"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration for all models.
// It creates tables, adds missing columns/indexes without destroying existing data.
func AutoMigrate(db *gorm.DB) error {
	log.Println("[database] running auto migration...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Task{},
		&models.Plugin{},
		&models.Workflow{},
	)
	if err != nil {
		return err
	}

	log.Println("[database] auto migration completed successfully")
	return nil
}