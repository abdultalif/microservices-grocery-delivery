package repository

import (
	"context"
	"testing"
	"user-service/database/seeds"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	dbConnString := "host=localhost port=5432 user=postgres password=lokal dbname=test sslmode=disable"
	return gorm.Open(postgres.Open(dbConnString), &gorm.Config{})
}

func truncateTables(db *gorm.DB) {
	db.Exec("DELETE FROM verification_tokens")
	db.Exec("DELETE FROM user_roles")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM roles")
}

func TestCreateUserAccountSuccess(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	
	truncateTables(db)
	seeds.SeedRoles(db)

	userRepo := NewUserRepository(db)

	user := entity.UserEntity{
		Name:     "Abdul Talif",
		Email:    "talif@gmail.com",
		Password: "admin123",

	}

	err = userRepo.CreateUserAccount(context.Background(), user)
	assert.NoError(t, err)

	var modelUser  model.User
	err = db.Where("email = ?", user.Email).First(&modelUser ).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Name, modelUser .Name)
}

func TestCreateUserAccountFailedMissingRole(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	
	truncateTables(db)

	userRepo := NewUserRepository(db)

	user := entity.UserEntity{
		Name:     "Test User",
		Email:    "failuser@gmail.com",
		Password: "123456",
	}

	err = userRepo.CreateUserAccount(context.Background(), user)
	assert.NoError(t, err)
}


func TestFindUserByEmail(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	
	truncateTables(db)
	seeds.SeedRoles(db)

	userRepo := NewUserRepository(db)

	user := entity.UserEntity{
		Name:     "Abdul Talif",
		Email:    "talif@gmail.com",
		Password: "admin123",
	}

	_ = userRepo.CreateUserAccount(context.Background(), user)
	assert.NoError(t, err)

	found, err := userRepo.FindUserByEmail(context.Background(), user.Email)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.Name, found.Name)
}

func TestFindUserByEmailNotFound(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	truncateTables(db)	

	userRepo := NewUserRepository(db)

	result, err := userRepo.FindUserByEmail(context.Background(), "notfound@gmail.com")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCreateUserAccountDuplicateEmail(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	truncateTables(db)
	seeds.SeedRoles(db)

	userRepo := NewUserRepository(db)

	user := entity.UserEntity{
		Name:     "Abdul",
		Email:    "abdultalif@gmail.com",
		Password: "secret",
	}
	err = userRepo.CreateUserAccount(context.Background(), user)
	assert.NoError(t, err)

	user2 := entity.UserEntity{
		Name:     "Talif",
		Email:    "abdultalif@gmail.com", // sengajan menggunakan email yang sama
		Password: "secret",
	}
	err = userRepo.CreateUserAccount(context.Background(), user2)
	assert.Error(t, err)
}

