package storage

import (
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("[storage] running auto migration...")
	err := db.AutoMigrate(
		&User{},
		&Session{},
		&APIKey{},
		&Permission{},
		&Quota{},
		&Project{},
		&Plugin{},
		&Task{},
		&TaskLog{},
	)
	if err != nil {
		return err
	}
	log.Println("[storage] auto migration completed successfully")
	return nil
}
