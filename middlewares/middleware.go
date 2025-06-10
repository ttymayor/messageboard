package middleware

import (
	"errors"
	"messageboard/models"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getAllowedDomains() []string {
	domains := os.Getenv("ALLOWED_DOMAINS")
	return strings.Split(domains, ",")
}

// DomainRestriction 檢查請求來源是否為允許的域名
func DomainRestriction() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得允許的域名列表
		allowedDomains := getAllowedDomains()
		if len(allowedDomains) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "未設定允許的域名"})
			c.Abort()
			return
		}

		// 檢查 Origin 或 Referer 頭
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")

		// 優先使用 Origin，如果沒有再使用 Referer
		source := origin
		if source == "" {
			source = referer
		}

		// 如果都沒有，拒絕請求
		if source == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法確認請求來源"})
			c.Abort()
			return
		}

		// 解析 URL 獲取域名
		parsedURL, err := url.Parse(source)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "無效的請求來源"})
			c.Abort()
			return
		}

		// 檢查域名是否在允許列表中
		hostname := parsedURL.Hostname()
		hostWithPort := parsedURL.Host
		allowed := false

		for _, domain := range allowedDomains {
			// 檢查完整域名匹配
			if domain == hostname || domain == hostWithPort {
				allowed = true
				break
			}

			// 檢查子域名匹配（允許 *.example.com）
			if strings.HasPrefix(domain, "*.") {
				baseDomain := domain[2:] // 去掉 "*."
				if strings.HasSuffix(hostname, baseDomain) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "你似乎想 CSRF 攻擊我的後端ㄟ。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 身分驗中介軟體，使用 JWT 進行授權
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得 Authorization 標頭
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授權資訊"})
			c.Abort()
			return
		}

		// 取得 Token 字串
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析 JWT Token
		token, err := jwt.ParseWithClaims(tokenString, &models.AppClaims{}, func(token *jwt.Token) (interface{}, error) { // 使用 ParseWithClaims 和 struct 指標
			// 驗證簽名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid // 或更明確的錯誤
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// 檢查解析錯誤和 Token 有效性
		if err != nil {
			// Use errors.Is for specific validation errors in v5
			if errors.Is(err, jwt.ErrTokenMalformed) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式錯誤"})
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) { // Check signature invalidity specifically
				c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的簽名"})
			} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 已過期或尚未生效"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "無法處理 Token: " + err.Error()})
			}
			c.Abort()
			return
		}

		// 驗證 Token 的 Claims
		if claims, ok := token.Claims.(*models.AppClaims); ok && token.Valid { // 斷言為 *models.AppClaims
			// 取得使用者 ID，並查詢使用者資料
			userID := claims.UserID // 直接從 struct 讀取，型別安全
			var user models.User
			if err := models.DB.First(&user, userID).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用戶不存在"})
				c.Abort()
				return
			}
			// 更新使用者的最後登入時間
			user.LastLogin = time.Now()
			if err := models.DB.Save(&user).Error; err != nil {
				// 注意：這裡記錄錯誤可能比直接回傳 500 更好，避免影響主要流程
				// log.Printf("更新最後登入時間失敗: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新最後登入時間失敗"})
				c.Abort()
				return
			}

			// 儲存至 context
			c.Set("currentUser", user)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的 Token Claims"})
			c.Abort()
		}
	}
}
