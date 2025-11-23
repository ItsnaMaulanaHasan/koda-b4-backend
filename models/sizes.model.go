package models

import (
	"backend-daily-greens/config"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Sizes struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" form:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy int       `json:"created_by,omitempty" db:"-"`
	UpdatedBy int       `json:"updated_by,omitempty" db:"-"`
}

func GetTotalDataSizes(search string) (int, error) {
	totalData := 0
	var err error
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM sizes WHERE name ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM sizes`).Scan(&totalData)
	}
	if err != nil {
		return totalData, err
	}

	return totalData, nil
}

func GetListAllSizes(page int, limit int, search string) ([]Sizes, string, error) {
	offset := (page - 1) * limit
	var rows pgx.Rows
	var err error
	message := ""
	sizes := []Sizes{}

	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM sizes
			WHERE name ILIKE $3
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM sizes
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		message = "Failed to fetch sizes from database"
		return sizes, message, err
	}
	defer rows.Close()

	sizes, err = pgx.CollectRows(rows, pgx.RowToStructByName[Sizes])
	if err != nil {
		message = "Failed to process size data from database"
		return sizes, message, err
	}

	message = "Success get all sizes"
	return sizes, message, nil
}
