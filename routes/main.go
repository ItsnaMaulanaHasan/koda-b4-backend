package routes

import (
	"backend-daily-greens/middlewares"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(r *gin.Engine) {
	authRouter(r.Group("/auth"))
	cartsRouter(r.Group("/carts", middlewares.Auth()))
	profilesRoutes(r.Group("/profiles", middlewares.Auth()))
	historiesRoutes(r.Group("/histories", middlewares.Auth()))

	admin := r.Group("/admin", middlewares.Auth(), middlewares.AdminOnly())
	usersRoutes(admin)
	categoriesRoutes(admin)
	productsRoutes(r, admin)
	transactionsRoutes(r, admin)
}
