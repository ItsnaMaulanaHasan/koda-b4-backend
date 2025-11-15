package lib

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserPayload struct {
	Id        int    `json:"id"`
	Role      string `json:"role"`
	SessionId int    `json:"sessionId"`
	jwt.RegisteredClaims
}

func GenerateToken(id int, role string, sessionId int) (string, error) {
	secretKey := []byte(os.Getenv("APP_SECRET"))
	claims := UserPayload{
		id,
		role,
		sessionId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(secretKey)
	return ss, err
}
