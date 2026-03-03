package otelgorm

import (
	"context"
	"testing"

	"github.com/QuickBill-Engineering/qbp-lib/tracing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestUser struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestNewPlugin_Basic(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)

	plugin := NewPlugin()
	err = db.Use(plugin)
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}
}

func TestNewPlugin_WithOptions(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)

	plugin := NewPlugin(
		WithDBName("testdb"),
		WithLogQueries(true),
		WithRecordRowsAffected(true),
	)
	err = db.Use(plugin)
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}
}

func TestPlugin_Create(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(WithDBName("testdb")))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	user := &TestUser{Name: "John Doe"}
	result := db.Create(user)
	if result.Error != nil {
		t.Errorf("failed to create user: %v", result.Error)
	}
	if user.ID == 0 {
		t.Error("expected user ID to be set after creation")
	}
}

func TestPlugin_Query(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(WithDBName("testdb")))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	db.Create(&TestUser{Name: "John Doe"})

	var user TestUser
	result := db.First(&user, 1)
	if result.Error != nil {
		t.Errorf("failed to query user: %v", result.Error)
	}
	if user.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", user.Name)
	}
}

func TestPlugin_Update(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(WithDBName("testdb")))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	user := &TestUser{Name: "John Doe"}
	db.Create(user)

	user.Name = "Jane Doe"
	result := db.Save(user)
	if result.Error != nil {
		t.Errorf("failed to update user: %v", result.Error)
	}

	var updatedUser TestUser
	db.First(&updatedUser, user.ID)
	if updatedUser.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %s", updatedUser.Name)
	}
}

func TestPlugin_Delete(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	err = db.Use(NewPlugin(WithDBName("testdb")))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	db.Exec("DELETE FROM test_users")

	user := &TestUser{Name: "John Doe"}
	db.Create(user)

	result := db.Delete(user)
	if result.Error != nil {
		t.Errorf("failed to delete user: %v", result.Error)
	}

	var count int64
	db.Model(&TestUser{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 users after delete, got %d", count)
	}
}

func TestPlugin_ExcludeTables(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(
		WithDBName("testdb"),
		WithExcludeTables("test_users"),
	))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	user := &TestUser{Name: "John Doe"}
	result := db.Create(user)
	if result.Error != nil {
		t.Errorf("failed to create user: %v", result.Error)
	}
}

func TestPlugin_WithContext(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(WithDBName("testdb")))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	ctx := context.Background()
	user := &TestUser{Name: "John Doe"}
	result := db.WithContext(ctx).Create(user)
	if result.Error != nil {
		t.Errorf("failed to create user with context: %v", result.Error)
	}
}

func TestPlugin_RecordRowsAffected(t *testing.T) {
	shutdown, err := tracing.Init(tracing.WithEnabled(false))
	if err != nil {
		t.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(nil)

	db := setupTestDB(t)
	err = db.Use(NewPlugin(
		WithDBName("testdb"),
		WithRecordRowsAffected(true),
	))
	if err != nil {
		t.Fatalf("failed to use plugin: %v", err)
	}

	user := &TestUser{Name: "John Doe"}
	result := db.Create(user)
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}
}
