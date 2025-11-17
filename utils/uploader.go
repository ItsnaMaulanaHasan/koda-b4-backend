package utils

import (
	"backend-daily-greens/config"
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file multipart.File, filePath string) (string, error) {
	cld, err := config.Cloudinary()
	if err != nil {
		return "", err
	}

	uploadParams := uploader.UploadParams{
		PublicID: filePath,
	}

	result, err := cld.Upload.Upload(context.Background(), file, uploadParams)
	if err != nil {
		return "", err
	}

	imageUrl := result.SecureURL
	return imageUrl, nil
}
