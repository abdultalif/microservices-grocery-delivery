package test

import (
	"context"
	"testing"
	"user-service/internal/adapter/repository"
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

func TestCreateUserAccount(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Exec("DELETE FROM users")

	userRepo := repository.NewUserRepository(db)

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

func TestFindUserByEmail(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Exec("DELETE FROM users")

	userRepo := repository.NewUserRepository(db)

	user := entity.UserEntity{
		Name:     "Abdul Talif",
		Email:    "talif@gmail.com",
		Password: "admin123",
	}

	err = userRepo.CreateUserAccount(context.Background(), user)
	assert.NoError(t, err)

	foundUser , err := userRepo.FindUserByEmail(context.Background(), user.Email)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser )
	assert.Equal(t, user.Name, foundUser .Name)
}
