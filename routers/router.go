package routers

import (
	"messageboard/controllers"
	middleware "messageboard/middlewares"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getAllowedOrigins() []string {
	// 從環境變數讀取允許的來源
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		// 開發環境預設值
		return []string{"http://localhost:3000", "http://localhost:8080"}
	}
	return strings.Split(origins, ",")
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 配置 CORS 中介軟體
	r.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// IP 限流
	r.Use(middleware.RateLimitPerIP())

	api := r.Group("/api")
	v1 := api.Group("/v1")

	// Public routes
	v1.POST("/register", controllers.Register)
	v1.POST("/login", controllers.Login)

	// Public comment routes (不需要認證)
	publicComments := v1.Group("/comments")
	{
		publicComments.GET("", controllers.GetComments)               // GET /api/v1/comments/
		publicComments.GET("/by-url", controllers.GetCommentsByURL)   // GET /api/v1/comments/by-url?url=xxx
		publicComments.GET("/:id", controllers.GetCommentByID)        // GET /api/v1/comments/:id
		publicComments.GET("/:id/likes", controllers.GetCommentLikes) // GET /api/v1/comments/:id/likes
	}

	// Protected routes (需要認證)
	authGroup := v1.Group("/")
	authGroup.Use(middleware.JWTAuth())

	// Protected comment routes (需要認證的寫入操作)
	protectedComments := authGroup.Group("/comments")
	{
		protectedComments.POST("", controllers.CreateComment)              // POST /api/v1/comments/
		protectedComments.PUT("/:id", controllers.UpdateComment)           // PUT /api/v1/comments/:id
		protectedComments.DELETE("/:id", controllers.DeleteComment)        // DELETE /api/v1/comments/:id
		protectedComments.POST("/:id/like", controllers.ToggleCommentLike) // POST /api/v1/comments/:id/like
	}

	// Test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	return r
}
