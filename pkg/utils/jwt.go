package utils

import (
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/domain"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(user *domain.User, cfg *config.Config) (string, error) {
	claims := JWTClaims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWTIssuer,
			Subject:   strconv.Itoa(int(user.ID)),
			Audience:  []string{cfg.JWTAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTExpire)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret
	signedToken, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString string, jwtSecret string) (string, error) {
	panic("TODO: implement ValidateJWT")
}
