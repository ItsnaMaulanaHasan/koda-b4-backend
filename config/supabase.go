package config

import (
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

var StorageClient *storage_go.Client

func InitSupabase() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	StorageClient = storage_go.NewClient(
		supabaseURL+"/storage/v1",
		supabaseKey,
		nil,
	)
}
