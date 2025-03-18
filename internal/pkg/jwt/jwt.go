package jwtToken

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func New(
	uid string,
	isAdmin bool,
	tokenTTL time.Duration,
	secret []byte,
) (
	string,
	error,
) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = uid
	claims["is_admin"] = isAdmin
	claims["exp"] = time.Now().Add(tokenTTL).Unix()

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string, secret []byte) (string, bool, error) {
	// Parse the token with the secret key
	//token, err := jwt.Parse(
	//	tokenString, func(token *jwt.Token) (interface{}, error) {
	//		return secret, nil
	//	},
	//)

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
	)

	// Check for verification errors
	if err != nil {
		return "", false, err
	}

	// Check if the token is valid
	if !token.Valid {
		return "", false, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return claims["uid"].(string), claims["is_admin"].(bool), nil
}
