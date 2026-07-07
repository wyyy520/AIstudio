package database

import (
	"os"
	"sync"
	"testing"

	"github.com/aistudio/backend/internal/database/models"
)

func TestLoadConfig(t *testing.T) {
	// Use SQLite for tests
	os.Setenv("DATABASE_TYPE", "sqlite")
	os.Setenv("DATABASE_URL", "file::memory:?cache=shared")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if cfg.Type != "sqlite" {
		t.Errorf("expected sqlite, got %s", cfg.Type)
	}
}

func TestInitAndMigrate(t *testing.T) {
	os.Setenv("DATABASE_TYPE", "sqlite")
	os.Setenv("DATABASE_URL", "file::memory:?cache=shared")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Reset once for testing
	once = sync.Once{}
	db = nil

	if err := Init(cfg); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	if !IsReady() {
		t.Fatal("IsReady() should be true after Init()")
	}

	// Verify tables exist
	database := GetDB()
	if !database.Migrator().HasTable(&models.User{}) {
		t.Error("users table not created")
	}
	if !database.Migrator().HasTable(&models.Project{}) {
		t.Error("projects table not created")
	}
	if !database.Migrator().HasTable(&models.Task{}) {
		t.Error("tasks table not created")
	}
	if !database.Migrator().HasTable(&models.Plugin{}) {
		t.Error("plugins table not created")
	}
	if !database.Migrator().HasTable(&models.Workflow{}) {
		t.Error("workflows table not created")
	}
}

func TestCRUD(t *testing.T) {
	os.Setenv("DATABASE_TYPE", "sqlite")
	os.Setenv("DATABASE_URL", "file::memory:?cache=shared")

	cfg, _ := LoadConfig()
	once = sync.Once{}
	db = nil

	if err := Init(cfg); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	database := GetDB()

	// Create
	user := models.User{Username: "testuser", Email: "test@example.com", Password: "secret"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("Create user failed: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("user ID should not be 0 after creation")
	}

	// Read
	var found models.User
	if err := database.First(&found, user.ID).Error; err != nil {
		t.Fatalf("First() failed: %v", err)
	}
	if found.Username != "testuser" {
		t.Errorf("expected testuser, got %s", found.Username)
	}

	// Update
	if err := database.Model(&found).Update("username", "updateduser").Error; err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	database.First(&found, user.ID)
	if found.Username != "updateduser" {
		t.Errorf("expected updateduser, got %s", found.Username)
	}

	// Delete
	if err := database.Delete(&found).Error; err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
	if err := database.First(&models.User{}, user.ID).Error; err == nil {
		t.Error("user should be deleted")
	}
}