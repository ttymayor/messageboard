package routers

import (
	"messageboard/controllers"
	middleware "messageboard/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 啟用 CORS 中介軟體
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 或指定域名
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	r.Use(middleware.DomainRestriction())
	v1 := api.Group("/v1")

	// Public routes
	v1.POST("/register", controllers.Register)
	v1.POST("/login", controllers.Login)

	// Protected routes
	authGroup := v1.Group("/")
	authGroup.Use(middleware.JWTAuth())

	// Comment routes
	comments := authGroup.Group("/comments")
	{
		comments.GET("", controllers.GetComments)             // GET /api/v1/comments/
		comments.POST("", controllers.CreateComment)          // POST /api/v1/comments/
		comments.GET("/by-url", controllers.GetCommentsByURL) // GET /api/v1/comments/by-url?url=xxx
		comments.GET("/:id", controllers.GetCommentByID)      // GET /api/v1/comments/:id
		comments.PUT("/:id", controllers.UpdateComment)       // PUT /api/v1/comments/:id
		comments.DELETE("/:id", controllers.DeleteComment)    // DELETE /api/v1/comments/:id

		// Like routes
		comments.POST("/:id/like", controllers.ToggleCommentLike) // POST /api/v1/comments/:id/like
		comments.GET("/:id/likes", controllers.GetCommentLikes)   // GET /api/v1/comments/:id/likes
	}

	// Test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	return r
}
