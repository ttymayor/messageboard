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
	authGroup.Use(middleware.JWTAuth())

	// Comment routes
	authGroup.GET("/comments", controllers.GetComments)
	authGroup.POST("/comments", controllers.CreateComment)
	authGroup.GET("/comments/:id", controllers.GetCommentByID)
	authGroup.DELETE("/comments/:id", controllers.DeleteComment)

	// Like routes
	authGroup.POST("/comments/:id/like", controllers.ToggleCommentLike)
	authGroup.GET("/comments/:id/likes", controllers.GetCommentLikes)

	// Get comments by URL
	authGroup.GET("/comments/:url", controllers.GetCommentsByURL)

	// Test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	return r
}
