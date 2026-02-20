package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
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

func TestJWT(t *testing.T) {
	randUUID, _ := uuid.NewRandom()

	token, _ := MakeJWT(randUUID, "secret", time.Second*1)
	returnedUUID, _ := ValidateJWT(token, "secret")

	if returnedUUID != randUUID {
		t.Errorf("UUID mismatch")
	}

	token, _ = MakeJWT(randUUID, "secret", time.Second*1)

	time.Sleep(2 * time.Second)

	returnedUUID, err := ValidateJWT(token, "secret")
	if err == nil {
		t.Errorf("Expected JWT rejection due to timeout")
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{}
	header.Add("Authorization", "Bearer token")

	if value, _ := GetBearerToken(header); value != "token" {
		t.Errorf("Token mismatch")
	}

	header = http.Header{}
	if _, err := GetBearerToken(header); err == nil {
		t.Errorf("Expected error due to absence of the token")
	}
}
