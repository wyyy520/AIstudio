package storage

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	mu   sync.RWMutex
	once sync.Once
)

type Config struct {
	Type string
	URL  string
}

func (c *Config) DSN() string {
	return c.URL
}

func Init(cfg *Config) error {
	var initErr error
	once.Do(func() {
		log.Printf("[storage] initializing %s database: %s", cfg.Type, cfg.URL)
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
		sqlDB, err := gormDB.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get underlying sql.DB: %w", err)
			return
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		db = gormDB
		if err := AutoMigrate(db); err != nil {
			initErr = fmt.Errorf("auto migration failed: %w", err)
			return
		}
		log.Printf("[storage] %s database initialized successfully", cfg.Type)
	})
	return initErr
}

func GetDB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	if db == nil {
		log.Fatal("[storage] database not initialized. Call storage.Init() first.")
	}
	return db
}

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
	log.Println("[storage] closing database connection...")
	return sqlDB.Close()
}

func IsReady() bool {
	mu.RLock()
	defer mu.RUnlock()
	return db != nil
}
