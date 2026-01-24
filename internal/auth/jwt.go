package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"userID":     userID,
		"expiryDate": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func VerifyToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) { return secret, nil })
	if err != nil || !token.Valid {
		return 0, errors.New("Invalid token.")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("Invalid claims.")
	}
	uid, ok := claims["userID"].(float64)
	if !ok {
		return 0, errors.New("Invalid userID.")
	}
	return int(uid), nil
}
