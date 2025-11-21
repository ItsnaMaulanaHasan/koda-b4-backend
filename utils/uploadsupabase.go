package utils

import (
	"backend-daily-greens/config"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

func UploadToSupabase(file *multipart.FileHeader, fileName string, bucketName string) (string, error) {
	// Open file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Get file extension
	ext := filepath.Ext(file.Filename)
	fullFileName := fileName + ext

	// Upload to Supabase
	_, err = config.StorageClient.UploadFile(bucketName, fullFileName, src)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	// Generate public URL
	supabaseURL := os.Getenv("SUPABASE_URL")
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		supabaseURL,
		bucketName,
		fullFileName,
	)

	return publicURL, nil
}
