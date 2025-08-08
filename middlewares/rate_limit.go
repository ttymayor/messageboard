package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

var clients = make(map[string]*Client)
var mu sync.Mutex

func getClient(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if c, exists := clients[ip]; exists {
		c.LastSeen = time.Now()
		return c.Limiter
	}

	// 允許每秒 1 次，突發 3 次
	limiter := rate.NewLimiter(1, 3)
	clients[ip] = &Client{Limiter: limiter, LastSeen: time.Now()}
	return limiter
}

func RateLimitPerIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getClient(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "我不允許你 DDoS 我"})
			c.Abort()
			return
		}
		c.Next()
	}
}
