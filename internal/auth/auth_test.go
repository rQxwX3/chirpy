package auth

import (
	"testing"
)

func TestHashing(t *testing.T) {
	passwords := []string{"", "hello", "123hello", "123hello!"}

	for _, password := range passwords {
		hash, _ := HashPassword(password)

		if ok, _ := CheckPasswordHash(password, hash); !ok {
			t.Errorf("Hashes do not match")
		}
	}
}
