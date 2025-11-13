package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type UserProfile struct {
	Id           int    `db:"id" json:"id"`
	ProfilePhoto string `db:"profile_photo" json:"profilePhoto"`
	FullName     string `db:"full_name" json:"fullName"`
	Email        string `db:"email" json:"email"`
	Phone        string `db:"phone" json:"phone"`
	Address      string `db:"address" json:"address"`
	Password     string `db:"password" json:"-"`
}

func GetProfileById(userId int) (UserProfile, string, error) {
	user := UserProfile{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			COALESCE(p.image, '') AS profile_photo,
			(u.first_name || ' ' || u.last_name) AS full_name,
			u.email,
			COALESCE(p.address, '') AS address,
			COALESCE(p.phone, '') AS phone,
			u.password
		FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE u.id = $1`, userId)
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
