package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, nil)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	ok, _, err := argon2id.CheckHash(password, hash)
	if err != nil {
		return false, err
	}

	return ok, nil
}
