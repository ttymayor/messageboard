package routers

import (
	"messageboard/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	// Set up routes for comments
	v1.POST("/comments", controllers.CreateComment)
	v1.GET("/comments", controllers.GetComments)
	v1.DELETE("/comments/:id", controllers.DeleteComment)
	// v1.GET("/comments/:id", controllers.GetCommentByID)
	// v1.PUT("/comments/:id", controllers.UpdateComment)

	// Set up routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	return r
}
