package database

import (
	"log"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("[database] running auto migration...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Task{},
		&models.Plugin{},
		&models.Workflow{},
		&models.Session{},
		&models.APIKey{},
		&models.Permission{},
		&models.Quota{},
	)
	if err != nil {
		return err
	}

	log.Println("[database] auto migration completed successfully")
	return nil
}
