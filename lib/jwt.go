package lib

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type UserPayload struct {
	Id int `json:"id"`
	jwt.RegisteredClaims
}

func GenerateToken(id int) (string, error) {
	godotenv.Load()
	secretKey := []byte(os.Getenv("APP_SECRET"))
	claims := UserPayload{
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(secretKey)
	return ss, err
}
