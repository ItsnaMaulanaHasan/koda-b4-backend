package models

type User struct {
	Id        int    `form:"id" swaggerignore:"true"`
	FirstName string `form:"first_name" binding:"required,min=3,max=20" example:"koda"`
	LastName  string `form:"last_name" binding:"required,min=3,max=20" example:"koda"`
	Email     string `form:"email" binding:"required,email" example:"koda@mail.com"`
	Role      string `form:"role" swaggerignore:"true" example:"customer/admin"`
	Password  string `form:"password" binding:"required,min=6" example:"koda123" format:"password"`
}
