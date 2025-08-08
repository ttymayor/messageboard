package middlewares

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ClientMap struct {
	m sync.Map
}

// Get returns the client from the map
func (m *ClientMap) Get(ip string) *Client {
	if client, exists := m.m.Load(ip); exists {
		return client.(*Client)
	}
	return nil
}

// Set adds a client to the map
func (m *ClientMap) Set(ip string, client *Client) {
	m.m.Store(ip, client)
}

// Delete removes the client from the map
func (m *ClientMap) Delete(ip string) {
	m.m.Delete(ip)
}

// Gc removes the expired clients that have not been seen in the last 10 minutes
func (m *ClientMap) Gc() {
	m.m.Range(func(key, value interface{}) bool {
		if time.Since(value.(*Client).LastSeen) > 10*time.Minute {
			m.m.Delete(key)
		}
		return true
	})
}

var clientMap = &ClientMap{}

type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

func getClient(ip string) *rate.Limiter {
	if limiter := clientMap.Get(ip); limiter != nil {
		limiter.LastSeen = time.Now()
		clientMap.Set(ip, limiter)
		return limiter.Limiter
	}

	// 允許每秒 1 次，突發 3 次
	limiter := rate.NewLimiter(1, 3)
	clientMap.Set(ip, &Client{Limiter: limiter, LastSeen: time.Now()})
	return limiter
}

func RateLimitPerIP() gin.HandlerFunc {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("正在從 map 清理 10 分鐘沒有見到的 clients")
			clientMap.Gc()
		}
	}()

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
