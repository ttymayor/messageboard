package middlewares

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
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
	m.m.Range(func(key, value any) bool {
		if time.Since(time.Unix(value.(*Client).LastSeen.Load(), 0)) > 10*time.Minute {
			m.m.Delete(key)
		}
		return true
	})
}

var clientMap = &ClientMap{}

type Client struct {
	Limiter  *rate.Limiter
	LastSeen *atomic.Int64
}

func getClient(ip string) *rate.Limiter {
	if limiter := clientMap.Get(ip); limiter != nil {
		limiter.LastSeen.Store(time.Now().Unix())
		return limiter.Limiter
	}

	// 允許每秒 1 次，突發 3 次
	limiter := rate.NewLimiter(1, 3)

	lastSeen := &atomic.Int64{}
	lastSeen.Store(time.Now().Unix())

	clientMap.Set(ip, &Client{Limiter: limiter, LastSeen: lastSeen})
	return limiter
}

var cleanerGoroutine sync.Once

func RateLimitPerIP() gin.HandlerFunc {
	cleanerGoroutine.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				log.Println("正在從 map 清理 10 分鐘沒有見到的 clients")
				clientMap.Gc()
			}
		}()
	})

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
