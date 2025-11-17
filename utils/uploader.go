package utils

import (
	"backend-daily-greens/config"
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file *multipart.FileHeader, filePath string) (string, error) {
	cld, err := config.Cloudinary()
	if err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	uploadParams := uploader.UploadParams{
		PublicID: filePath,
		Folder:   "products",
	}

	result, err := cld.Upload.Upload(context.Background(), src, uploadParams)
	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}
