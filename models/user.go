package models

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

type RegisterUserRequest struct {
	Id        int    `swaggerignore:"true"`
	FirstName string `form:"first_name" example:"John"`
	LastName  string `form:"last_name" example:"Cena"`
	Email     string `form:"email" example:"johncena@mail.com"`
	Password  string `form:"password" example:"koda123" format:"password"`
	Role      string `form:"role" example:"customer/admin"`
}

type UpdateUserRequest struct {
	ProfilePhoto string `form:"profilephoto"`
	FirstName    string `form:"first_name" example:"John"`
	LastName     string `form:"last_name" example:"Cena"`
	Phone        string `form:"phone" example:"0895367608879"`
	Address      string `form:"address" example:"jl. sadewa 1"`
	Email        string `form:"email" example:"johncena@mail.com"`
	Role         string `form:"role" example:"customer/admin"`
}

type UserResponse struct {
	Id           int     `json:"id" db:"id"`
	ProfilePhoto *string `json:"profilephoto" db:"image"`
	FirstName    string  `json:"first_name" db:"first_name"`
	LastName     string  `json:"last_name" db:"last_name"`
	Phone        *string `json:"phone" db:"phone_number"`
	Address      *string `json:"address" db:"address"`
	Email        string  `json:"email" db:"email"`
	Role         string  `json:"role" db:"role"`
}

var ResponseUserData *UserResponse
