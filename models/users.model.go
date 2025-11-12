package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Id           int
	ProfilePhoto string `form:"profilephoto"`
	FirstName    string `form:"first_name" example:"John"`
	LastName     string `form:"last_name" example:"Cena"`
	Phone        string `form:"phone" example:"0895367608879"`
	Address      string `form:"address" example:"jl. sadewa 1"`
	Email        string `form:"email" example:"johncena@mail.com"`
	Password     string `form:"password" example:"koda123" format:"password"`
	Role         string `form:"role" example:"customer/admin"`
}

type UserRegisterRequest struct {
	Id        int    `swaggerignore:"true"`
	FirstName string `form:"first_name" example:"John"`
	LastName  string `form:"last_name" example:"Cena"`
	Email     string `form:"email" example:"johncena@mail.com"`
	Password  string `form:"password" example:"koda123" format:"password"`
	Role      string `form:"role" example:"customer/admin"`
}

type UserUpdateRequest struct {
	ProfilePhoto string `form:"profilephoto"`
	FirstName    string `form:"first_name" example:"John"`
	LastName     string `form:"last_name" example:"Cena"`
	Phone        string `form:"phone" example:"0895367608879"`
	Address      string `form:"address" example:"jl. sadewa 1"`
	Email        string `form:"email" example:"johncena@mail.com"`
	Role         string `form:"role" example:"customer/admin"`
}

type UserResponse struct {
	Id           int    `json:"id" db:"id"`
	ProfilePhoto string `json:"profilephoto" db:"image"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	Phone        string `json:"phone" db:"phone_number"`
	Address      string `json:"address" db:"address"`
	Email        string `json:"email" db:"email"`
	Role         string `json:"role" db:"role"`
}

func GetUserById(id int) (UserResponse, string, error) {
	user := UserResponse{}
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

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[UserResponse])
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
