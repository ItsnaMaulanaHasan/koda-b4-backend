package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDatabase() {
	var err error
	ctx := context.Background()

	connStr := os.Getenv("DATABASE_URL")

	DB, err = pgxpool.New(ctx, connStr)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.Ping(ctx)
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully!")
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
