package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	Id           int                   `json:"id" db:"id"`
	FilePhoto    *multipart.FileHeader `json:"-" form:"filePhoto" db:"-"`
	ProfilePhoto string                `json:"profilePhoto" form:"-" db:"profile_photo"`
	FullName     string                `json:"fullName" form:"fullName" db:"full_name"`
	Phone        string                `json:"phone" form:"phone" db:"phone_number"`
	Address      string                `json:"address" form:"address" db:"address"`
	Email        string                `json:"email" form:"email" db:"email"`
	Password     string                `json:"-" form:"-" db:"-"`
	Role         string                `json:"role" form:"role" db:"role"`
}

func GetTotalDataUsers(search string) (int, error) {
	totalData := 0
	var err error
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM users 
			LEFT JOIN profiles ON users.id = profiles.user_id 
			WHERE profiles.full_name ILIKE $1
			OR profiles.phone_number ILIKE $1
			OR profiles.address ILIKE $1
			OR users.email ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&totalData)
	}
	if err != nil {
		return totalData, err
	}

	return totalData, err
}

func GetListAllUser(page int, limit int, search string) ([]User, string, error) {
	offset := (page - 1) * limit
	var rows pgx.Rows
	var err error
	message := ""
	users := []User{}
	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				users.id,
				COALESCE(profiles.profile_photo, '') AS profile_photo,
				COALESCE(profiles.full_name, '') AS full_name,
				COALESCE(profiles.phone_number, '') AS phone_number,
				COALESCE(profiles.address, '') AS address,
				users.email,
				users.role
			FROM users
			LEFT JOIN profiles ON users.id = profiles.user_id
			WHERE profiles.full_name ILIKE $3
			   OR profiles.phone_number ILIKE $3
			   OR profiles.address ILIKE $3
			   OR users.email ILIKE $3
			ORDER BY users.id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(
			context.Background(),
			`SELECT 
				users.id,
				COALESCE(profiles.profile_photo, '') AS profile_photo,
				COALESCE(profiles.full_name, '') AS full_name,
				COALESCE(profiles.phone_number, '') AS phone_number,
				COALESCE(profiles.address, '') AS address,
				users.email,
				users.role
			FROM users
			LEFT JOIN profiles ON users.id = profiles.user_id
			ORDER BY users.id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		message = "Failed to fetch users from database"
		return users, message, err
	}
	defer rows.Close()

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		message = "Failed to process user data from database"
		return users, message, err
	}

	message = "Success get all user"
	return users, message, nil
}

func GetDetailUser(id int) (User, string, error) {
	user := User{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			users.id,
			COALESCE(profiles.profile_photo, '') AS profile_photo,
			COALESCE(profiles.full_name, '') AS full_name,
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

func CheckUserEmail(email string) (bool, error) {
	exists := false
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email,
	).Scan(&exists)

	if err != nil {
		return exists, err
	}

	return exists, nil
}

func InsertDataUser(userId int, bodyCreate *User, filePatch string) (bool, string, error) {
	ctx := context.Background()
	isSuccess := false
	message := ""

	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	// insert into users
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, role, password, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		bodyCreate.Email,
		bodyCreate.Role,
		bodyCreate.Password,
		userId,
		userId,
	).Scan(&bodyCreate.Id)
	if err != nil {
		message = "Internal server error while inserting new user"
		return isSuccess, message, err
	}

	// insert into profiles
	_, err = tx.Exec(ctx,
		`INSERT INTO profiles (user_id, full_name, profile_photo, address, phone_number, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		bodyCreate.Id,
		bodyCreate.FullName,
		filePatch,
		bodyCreate.Address,
		bodyCreate.Phone,
		userId,
		userId,
	)
	if err != nil {
		message = "Internal server error while inserting new user profile"
		return isSuccess, message, err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "User created successfully"
	return isSuccess, message, nil
}

func UpdateDataUser(userId int, userIdFromToken int, bodyUpdate *User, savedFilePath string) (bool, string, error) {
	ctx := context.Background()
	isSuccess := false
	message := ""

	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		context.Background(),
		`UPDATE users 
		 SET full_name = COALESCE(NULLIF($1, ''), full_name),
		     email      = COALESCE(NULLIF($2, ''), email),
		     role       = COALESCE(NULLIF($3, ''), role),
		     updated_by = $4,
		     updated_at = NOW()
		 WHERE id = $5`,
		bodyUpdate.FullName,
		bodyUpdate.Email,
		bodyUpdate.Role,
		userIdFromToken,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user table"
		return isSuccess, message, err
	}

	_, err = tx.Exec(
		context.Background(),
		`UPDATE profiles 
		 SET profile_photo = COALESCE(NULLIF($1, ''), profile_photo),
		     address       = COALESCE(NULLIF($2, ''), address),
		     phone_number  = COALESCE(NULLIF($3, ''), phone_number),
		     updated_by    = $4,
		     updated_at    = NOW()
		 WHERE user_id = $5`,
		savedFilePath,
		bodyUpdate.Address,
		bodyUpdate.Phone,
		userIdFromToken,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user profile"
		return isSuccess, message, err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "User updated successfully"
	return isSuccess, message, nil
}

func DeleteDataUser(userId int) (pgconn.CommandTag, error) {
	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userId)
	if err != nil {
		return commandTag, err
	}

	return commandTag, err
}
