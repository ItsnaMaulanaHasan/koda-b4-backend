package utils

import (
	"backend-daily-greens/config"
	"context"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func DeleteFromCloudinary(imageUrl string) error {
	cld, err := config.Cloudinary()
	if err != nil {
		return err
	}

	parts := strings.Split(imageUrl, "/")
	filename := parts[len(parts)-1]
	publicID := strings.TrimSuffix(filename, filepath.Ext(filename))

	_, err = cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}
