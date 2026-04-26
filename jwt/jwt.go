package jwt

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func JWTDecrypt(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

type MyCustomClaims struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	jwt.RegisteredClaims
}

func JWTEncrypt(id string, myToken string, secret string) (string, error) {
	return JWTEncryptWithExpiry(id, myToken, secret, 24*time.Hour)
}

func JWTEncryptWithExpiry(id string, myToken string, secret string, expiry time.Duration) (string, error) {
	mySigningKey := []byte(secret)
	claims := MyCustomClaims{
		id,
		myToken,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			Issuer:    "admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}
