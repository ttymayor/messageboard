package middleware

import (
	"errors"
	"messageboard/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
