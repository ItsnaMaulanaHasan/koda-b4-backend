package models

type User struct {
	Id        int    `form:"id" swaggerignore:"true"`
	FirstName string `form:"first_name" example:"koda"`
	LastName  string `form:"last_name" example:"koda"`
	Email     string `form:"email" example:"koda@mail.com"`
}
