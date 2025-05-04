package controllers

import (
	"fmt"
	"log"
	"messageboard/models"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	mail "github.com/xhit/go-simple-mail/v2"
	"golang.org/x/crypto/bcrypt"
)

func CreateComment(c *gin.Context) {
	var input struct {
		URL      string `json:"url" binding:"required"` // 留言的網址
		Content  string `json:"content" binding:"required"`
		ParentID *uint  `json:"parent_id"` // 可選，若為 nil 則表示為根留言
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "參數錯誤"})
		return
	}

	// 驗證使用者是否登入
	userInterface, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "使用者未登入"})
		return
	}
	user := userInterface.(models.User)

	// 如果是回覆，確認父留言是否存在
	if input.ParentID != nil {
		var parentComment models.Comment
		if err := models.DB.First(&parentComment, *input.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "找不到要回覆的留言"})
			return
		}
	}

	// 檢查 URL 是否有效
	u, err := url.ParseRequestURI(input.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "網址格式錯誤"})
		return
	}

	// 建立留言
	comment := models.Comment{
		URL:      input.URL,
		ParentID: input.ParentID, // nil 表示主留言
		UserID:   user.ID,
		Content:  input.Content,
	}
	if err := models.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立留言失敗", "details": err.Error()})
		return
	}

	// 寄送通知信（可選）
	if err := sendEmailNotification(comment); err != nil {
		log.Printf("寄送通知信失敗: %v\n", err)
	} else {
		log.Printf("成功寄送通知信給 %s\n", comment.User.Username)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "留言成功",
		"comment": comment,
	})
}

func GetComments(c *gin.Context) {
	var comments []models.Comment
	if err := models.DB.Preload("User").Order("created_at DESC").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.Comment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

func GetCommentByID(c *gin.Context) {
	id := c.Param("id")
	var comment models.Comment
	if err := models.DB.Preload("User").First(&comment, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
}

/*
* Auth
*
* Register, Login
 */
func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=20"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤或欄位缺失"})
		return
	}

	// 檢查 email 是否已存在
	var existing models.User
	if err := models.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "此 Email 已被註冊"})
		return
	}

	// 密碼加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}

	newUser := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
		RoleID:   1, // 預設 Reader 角色
	}

	if err := models.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "註冊失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "註冊成功",
		"user": gin.H{
			"id":       newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
			"role_id":  newUser.RoleID,
		},
	})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "參數格式錯誤"})
		return
	}

	var user models.User
	if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "使用者不存在"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密碼錯誤"})
		return
	}

	// 產生 JWT Token
	// 設定 Token 的過期時間為 72 小時
	claims := models.AppClaims{ // 使用自訂 struct
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			// Issuer:    "your_app_name", // 可選
			// Subject:   strconv.FormatUint(uint64(user.ID), 10), // 可選
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 傳入 claims struct

	// 簽署 Token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "產生 token 失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登入成功",
		"token":   tokenString,
	})
}

func sendEmailNotification(comment models.Comment) error {
	server := mail.NewSMTPClient()
	server.Host = os.Getenv("MAIL_HOST")
	server.Port = getEnvAsInt("MAIL_PORT", 587)
	server.Username = os.Getenv("MAIL_USERNAME")
	server.Password = os.Getenv("MAIL_PASSWORD")
	server.Encryption = mail.EncryptionSTARTTLS
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	var toEmail string
	var subject string

	// 如果是回覆留言，通知父留言的作者
	if comment.ParentID != nil {
		var parentComment models.Comment
		if err := models.DB.Preload("User").First(&parentComment, *comment.ParentID).Error; err == nil && parentComment.User.Email != "" {
			toEmail = parentComment.User.Email
			subject = "【留言通知】你有一則新回覆"
		} else {
			// 找不到父留言或父留言作者沒信箱，通知自己
			toEmail = comment.User.Email
			subject = "【留言通知】你有一則新留言"
		}
	} else {
		// 主留言通知站長
		toEmail = os.Getenv("MAIL_TO")
		if toEmail == "" {
			toEmail = comment.User.Email
		}
		subject = "【留言通知】你有一則新留言"
	}

	email.SetFrom(os.Getenv("MAIL_FROM")).
		AddTo(toEmail).
		SetSubject(subject)

	body := fmt.Sprintf("```markdown\n## 作者：%s\n## 時間：%s\n## 內容：\n%s\n```",
		comment.User.Username,
		comment.CreatedAt.Format("2006-01-02 15:04:05"),
		comment.Content,
	)

	email.SetBody(mail.TextPlain, body)

	return email.Send(smtpClient)
}

func getEnvAsInt(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}
