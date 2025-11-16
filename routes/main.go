package routes

import (
	"backend-daily-greens/middlewares"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(r *gin.Engine) {
	authRouter(r.Group("/auth"))

	// admin
	admin := r.Group("/admin", middlewares.Auth(), middlewares.CheckSessionActive(), middlewares.AdminOnly())
	usersRoutes(admin)
	categoriesRoutes(admin)
	productsRoutes(r, admin)
	transactionsRoutes(r, admin)

	// public
	cartsRouter(r.Group("/carts", middlewares.Auth(), middlewares.CheckSessionActive()))
	profilesRoutes(r.Group("/profiles", middlewares.Auth(), middlewares.CheckSessionActive()))
	historiesRoutes(r.Group("/histories", middlewares.Auth(), middlewares.CheckSessionActive()))
	feeRoutes(r)
}
