package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "talif123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	valid := CheckPasswordHash(password, hash)
	assert.True(t, valid)

	invalid := CheckPasswordHash("wrongPassword", hash)
	assert.False(t, invalid)
}
