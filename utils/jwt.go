// utils/jwt.go
package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mfuadfakhruzzaki/backendaurauran/config"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
)

// Claims represents the JWT claims
type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID uint, role models.Role) (string, error) {
    expirationTime := time.Now().Add(time.Hour * 24) // 24 jam
    claims := &Claims{
        UserID: userID,
        Role:   string(role),
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
            IssuedAt:  time.Now().Unix(),
            Issuer:    "backendaurauran",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

// ParseJWT parses and validates a JWT token string
func ParseJWT(tokenStr string) (*Claims, error) {
    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        // Pastikan metode signing sesuai
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(config.AppConfig.JWT.Secret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
