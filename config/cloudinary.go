package config

import (
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

func Cloudinary() (*cloudinary.Cloudinary, error) {
	cldName := os.Getenv("CLOUDINARY_NAME")
	cldSecret := os.Getenv("CLOUDINARY_API_SECRET")
	cldKey := os.Getenv("CLOUDINARY_API_KEY")

	cld, err := cloudinary.NewFromParams(cldName, cldKey, cldSecret)
	if err != nil {
		return nil, err
	}

	return cld, nil
}
