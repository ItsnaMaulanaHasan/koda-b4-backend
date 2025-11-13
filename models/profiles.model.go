package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type UserProfile struct {
	Id           int    `db:""`
	ProfilePhoto string `db:"image"`
	FullName     string `db:"full_name"`
	Email        string `db:"email"`
	Phone        string `db:"phone"`
	Address      string `db:"address"`
	Password     string `db:"password"`
}

func GetProfileById(userId int) (UserProfile, string, error) {
	user := UserProfile{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			p.image
			(u.fisrt_name + u.last_name) AS full_name,
			u.email,
			p.address,
			p.phone
		FROM users u
		JOIN profiles p ON u.id = p.user_id
		WHERE u.id = $1)`, userId)
	if err != nil {
		message = "Internal server error while fetch data profile"
		return user, message, err
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[UserProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "User not found"
			return user, message, err
		}
		message = "Internal server error while process data profile"
		return user, message, err
	}

	message = "Success get profile user"
	return user, message, nil
}
