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


func TestLatLngToString(t *testing.T) {
	val := 106.84513
	str := LatLngToString(val)

	assert.Contains(t, str, "106")
	assert.NotEmpty(t, str)
}

func TestStringToInt64Success(t *testing.T) {
	input := "12345"
	expected := int64(12345)

	result, err := StringToInt64(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestStringToInt64Fail(t *testing.T) {
	input := "talif123"

	result, err := StringToInt64(input)
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
}
