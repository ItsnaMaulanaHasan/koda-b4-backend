package lib

import "github.com/matthewhartstonge/argon2"

func HashPassword(password string) (string, error) {
	config := argon2.DefaultConfig()

	hash, err := config.Hash([]byte(password), nil)
	if err != nil {
		return "", err
	}

	return string(hash.Encode()), nil
}
