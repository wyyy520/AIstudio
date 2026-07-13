package database

import (
	"fmt"
	"log"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	mu   sync.RWMutex
	once sync.Once
)

// Init initializes the database connection and runs auto-migration.
// Must be called once at application startup.
func Init(cfg *Config) error {
	var initErr error

	once.Do(func() {
		log.Printf("[database] initializing %s database: %s", cfg.Type, cfg.URL)

		var dialector gorm.Dialector
		switch cfg.Type {
		case "sqlite":
			dialector = sqlite.Open(cfg.DSN())
		case "postgres":
			dialector = postgres.Open(cfg.DSN())
		default:
			initErr = fmt.Errorf("unsupported database type: %s", cfg.Type)
			return
		}

		gormDB, err := gorm.Open(dialector, &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Configure connection pool
		sqlDB, err := gormDB.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get underlying sql.DB: %w", err)
			return
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		db = gormDB

		// Run auto-migration
		if err := AutoMigrate(db); err != nil {
			initErr = fmt.Errorf("auto migration failed: %w", err)
			return
		}

		log.Printf("[database] %s database initialized successfully", cfg.Type)
	})

	return initErr
}

// GetDB returns the initialized database instance.
// Panics if Init() has not been called yet.
func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		log.Fatal("[database] database not initialized. Call database.Init() first.")
	}
	return db
}

// Close gracefully closes the database connection.
func Close() error {
	mu.RLock()
	defer mu.RUnlock()

	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	log.Println("[database] closing database connection...")
	return sqlDB.Close()
}

// IsReady returns true if the database has been initialized.
func IsReady() bool {
	mu.RLock()
	defer mu.RUnlock()
	return db != nil
}