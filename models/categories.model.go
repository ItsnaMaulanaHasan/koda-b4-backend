package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Category struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" form:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy int       `json:"created_by,omitempty" db:"-"`
	UpdatedBy int       `json:"updated_by,omitempty" db:"-"`
}

func GetTotalDataCategories(search string) (int, error) {
	totalData := 0
	var err error
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM categories WHERE name ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM categories`).Scan(&totalData)
	}
	if err != nil {
		return totalData, err
	}

	return totalData, nil
}

func GetListAllCategories(page int, limit int, search string) ([]Category, string, error) {
	offset := (page - 1) * limit
	var rows pgx.Rows
	var err error
	message := ""
	categories := []Category{}

	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM categories
			WHERE name ILIKE $3
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM categories
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		message = "Failed to fetch categories from database"
		return categories, message, err
	}
	defer rows.Close()

	categories, err = pgx.CollectRows(rows, pgx.RowToStructByName[Category])
	if err != nil {
		message = "Failed to process category data from database"
		return categories, message, err
	}

	message = "Success get all categories"
	return categories, message, nil
}

func GetCategoryById(id int) (Category, string, error) {
	category := Category{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
		`SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id = $1`, id)
	if err != nil {
		message = "Failed to fetch category from database"
		return category, message, err
	}
	defer rows.Close()

	category, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Category])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Category not found"
			return category, message, err
		}
		message = "Failed to process category data"
		return category, message, err
	}

	message = "Success get category"
	return category, message, nil
}

func CheckCategoryName(name string) (bool, error) {
	var exists bool
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)", name,
	).Scan(&exists)

	if err != nil {
		return exists, err
	}

	return exists, nil
}

func InsertDataCategory(userId int, bodyCreate *Category) (bool, string, error) {
	isSuccess := false
	message := ""

	err := config.DB.QueryRow(
		context.Background(),
		`INSERT INTO categories (name, created_by, updated_by)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		bodyCreate.Name,
		userId,
		userId,
	).Scan(&bodyCreate.Id)
	if err != nil {
		message = "Internal server error while inserting new category"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "Category created successfully"
	return isSuccess, message, nil
}

func CheckCategoryNameExcludingId(name string, id int) (bool, error) {
	var exists bool
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND id != $2)", name, id,
	).Scan(&exists)

	if err != nil {
		return exists, err
	}

	return exists, nil
}

func UpdateDataCategory(categoryId int, userId int, name string) (bool, string, error) {
	isSuccess := false
	message := ""

	commandTag, err := config.DB.Exec(
		context.Background(),
		`UPDATE categories 
		 SET name = COALESCE(NULLIF($1, ''), name),
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		name,
		userId,
		categoryId,
	)
	if err != nil {
		message = "Internal server error while updating category"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Category not found"
		return isSuccess, message, nil
	}

	isSuccess = true
	message = "Category updated successfully"
	return isSuccess, message, nil
}

func DeleteDataCategory(categoryId int) (pgconn.CommandTag, error) {
	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM categories WHERE id = $1`, categoryId)
	if err != nil {
		return commandTag, err
	}

	return commandTag, nil
}
