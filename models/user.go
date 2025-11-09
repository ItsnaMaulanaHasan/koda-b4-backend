package models

type User struct {
	Id        int    `swaggerignore:"true"`
	FirstName string `form:"first_name" binding:"required" example:"John"`
	LastName  string `form:"last_name" binding:"required" example:"Cena"`
	Email     string `form:"email" binding:"required,email" example:"johncena@mail.com"`
	Role      string `form:"role" swaggerignore:"true"`
	Password  string `form:"password" binding:"required" example:"koda123" format:"password"`
}

type UpdateUserRequest struct {
	FirstName string `form:"first_name" binding:"required" example:"John"`
	LastName  string `form:"last_name" binding:"required" example:"Cena"`
	Email     string `form:"email" binding:"required,email" example:"johncena@mail.com"`
	Role      string `form:"role" example:"customer/admin"`
}

type UserResponse struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

var ResponseUserData *UserResponse
