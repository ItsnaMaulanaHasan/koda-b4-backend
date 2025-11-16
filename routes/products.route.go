package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func productsRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	products := admin.Group("/products")
	{
		products.GET("", controllers.ListProductsAdmin)
		products.GET("/:id", controllers.DetailProductAdmin)
		products.POST("", controllers.CreateProduct)
		products.PATCH("/:id", controllers.UpdateProduct)
		products.DELETE("/:id", controllers.DeleteProduct)

		products.GET(":id/images", controllers.ListProductImages)
		products.GET(":id/images/:imageId", controllers.DetailProductImage)
		products.POST(":id/images", controllers.CreateProductImage)
		products.PATCH(":id/images/:imageId", controllers.UpdateProductImage)
	}

	r.GET("/products", controllers.ListProductsPublic)
	r.GET("/products/:id", controllers.DetailProductPublic)
	r.GET("/favourite-products", controllers.ListFavouriteProducts)
}
