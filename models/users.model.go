package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Id           int    `json:"id" db:"id"`
	ProfilePhoto string `form:"profilePhoto" db:"image"`
	FirstName    string `form:"firstName" db:"first_name"`
	LastName     string `form:"lastName" db:"last_name"`
	Phone        string `form:"phone" db:"phone"`
	Address      string `form:"address" db:"address"`
	Email        string `form:"email" db:"email"`
	Password     string `form:"-" db:"-" json:"-"`
	Role         string `form:"role" db:"role"`
}

type UserProfile struct {
	Id           int `db:""`
	ProfilePhoto string
	FullName     string
	Email        string
	Phone        string
	Address      string
	Password     string
}

func GetUserById(id int) (User, string, error) {
	user := User{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			users.id,
			COALESCE(profiles.image, '') AS image,
			COALESCE(users.first_name, '') AS first_name,
			COALESCE(users.last_name, '') AS last_name,
			COALESCE(profiles.phone_number, '') AS phone_number,
			COALESCE(profiles.address, '') AS address,
			users.email,
			users.role
		FROM users
		LEFT JOIN profiles ON users.id = profiles.user_id
		WHERE users.id = $1`, id)
	if err != nil {
		message = "Failed to fetch user from database"
		return user, message, err
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "User not found"
			return user, message, err
		}
		message = "Failed to process user data"
		return user, message, err
	}

	message = "Success get user"
	return user, message, nil
}
