package config

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool
var once sync.Once

func InitDatabase() {
	once.Do(func() {
		ctx := context.Background()

		connStr := os.Getenv("DATABASE_URL")
		if connStr == "" {
			log.Fatal("DATABASE_URL is not set")
		}

		connStr = strings.Replace(connStr, "postgresql://", "postgres://", 1)

		var err error
		DB, err = pgxpool.New(ctx, connStr)

		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		err = DB.Ping(ctx)
		if err != nil {
			log.Fatal("Failed to ping database:", err)
		}

		log.Println("Database connected successfully!")
	})
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
