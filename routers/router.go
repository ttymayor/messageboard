package routers

import (
	"messageboard/controllers"
	middleware "messageboard/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	v1 := api.Group("/v1")

	// Public routes
	v1.POST("/register", controllers.Register)
	v1.POST("/login", controllers.Login)

	// Protected routes
	authGroup := v1.Group("/")
	authGroup.Use(middleware.JWTAuth()) // JWT 認證中介軟體
	authGroup.POST("/comments", controllers.CreateComment)
	authGroup.GET("/comments", controllers.GetComments)
	authGroup.DELETE("/comments/:id", controllers.DeleteComment)

	// Test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	return r
}
