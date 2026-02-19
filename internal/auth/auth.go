package auth

import (
	"errors"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentTime := time.Now().UTC()
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  &jwt.NumericDate{Time: currentTime},
		ExpiresAt: &jwt.NumericDate{Time: currentTime.Add(expiresIn)},
		Subject:   userID.String(),
	})

	signedJWT, err := jwt.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedJWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("Provided token is invalid")
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, err
	}

	return userUUID, nil
}
