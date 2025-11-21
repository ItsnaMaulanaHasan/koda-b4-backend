package utils

import (
	"backend-daily-greens/config"
	"fmt"
	"os"
	"strings"
)

func DeleteFromSupabase(imageUrl string, bucketName string) error {
	if imageUrl == "" {
		return nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")

	// Extract file path dari URL
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", supabaseURL, bucketName)
	filePath := strings.TrimPrefix(imageUrl, prefix)

	if filePath == imageUrl {
		return fmt.Errorf("invalid image URL format")
	}

	// Delete from Supabase
	_, err := config.StorageClient.RemoveFile(bucketName, []string{filePath})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}
