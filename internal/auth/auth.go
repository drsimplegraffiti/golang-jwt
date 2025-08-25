package auth

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

// Claims struct to include userID, username, and StandardClaims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID int64, username string, secretKey []byte) (string, error) {
	// Set claims
	claims := Claims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Expire in 24 hours
			Issuer:    "gotemp",                              // You can put the issuer name of your app here
		},
	}

	// Create token with claims and sign it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseJWT parses the JWT token and returns the claims
func ParseJWT(tokenString string, secretKey []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
