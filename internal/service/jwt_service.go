package service

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(phone_number string, admin bool) string
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type jwtCustomClaims struct {
	PhoneNumber string `json:"phone_number"`
	Admin       bool   `json:"admin"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService() JWTService {
	return &jwtService{
		secretKey: getSecretKey(),
		issuer:    "github.com/kimbasn/printly",
	}
}

func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

func (jwtSrv *jwtService) GenerateToken(phone_number string, admin bool) string {
	issueTime := time.Now()
	expirationTime := issueTime.Add(time.Hour * 72)

	claims := &jwtCustomClaims{
		phone_number,
		admin,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    jwtSrv.issuer,
			IssuedAt:  jwt.NewNumericDate(issueTime),
		},
	}

	// create token with claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded tokern using secret signing key
	tokenString, err := token.SignedString([]byte(jwtSrv.secretKey))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func (jwtSrv *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Siging method validation
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// return the secret signing key
		return []byte(jwtSrv.secretKey), nil
	})
}
