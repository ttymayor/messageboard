package controllers

import (
	"log"
	"messageboard/models"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	mail "github.com/xhit/go-simple-mail/v2"
)

/*
* Comment
*
* CreateComment, GetComments, DeleteComment, GetCommentByID, ToggleCommentLike
* 這些函數處理留言的建立、查詢、刪除和點讚功能
 */

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
	c.JSON(http.StatusOK, gin.H{
		"message":  "查詢成功",
		"comments": comments,
	})
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
	c.JSON(http.StatusOK, gin.H{
		"message": "查詢成功",
		"comment": comment,
	})
}

func GetCommentsByURL(c *gin.Context) {
	url := c.Param("url")
	var comments []models.Comment
	if err := models.DB.Where("url = ?", url).Preload("User").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "查詢成功",
		"comments": comments,
	})
}

func ToggleCommentLike(c *gin.Context) {
	commentID := c.Param("id")

	// 取得目前登入的使用者
	userInterface, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "使用者未登入"})
		return
	}
	user := userInterface.(models.User)

	// 檢查留言是否存在
	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "留言不存在"})
		return
	}

	// 檢查是否已經點過讚
	var existingLike models.CommentLike
	err := models.DB.
		Where("user_id = ? AND comment_id = ?", user.ID, comment.ID).
		First(&existingLike).Error

	if err == nil {
		// 已點過讚 → 取消讚
		if err := models.DB.Delete(&existingLike).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消讚失敗", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已取消讚"})
		return
	}

	// 未點過讚 → 新增讚
	newLike := models.CommentLike{
		UserID:    user.ID,
		CommentID: comment.ID,
	}
	if err := models.DB.Create(&newLike).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "點讚失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "點讚成功"})
}

func GetCommentLikes(c *gin.Context) {
	commentID := c.Param("id")

	// 檢查留言是否存在
	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "留言不存在"})
		return
	}

	var likes []models.CommentLike
	if err := models.DB.Where("comment_id = ?", comment.ID).Find(&likes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢點讚失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comment_id":  comment.ID,
		"likes_count": len(likes),
		"likes":       likes,
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

	// htmlBody := fmt.Sprintf("```markdown\n## 作者：%s\n## 時間：%s\n## 內容：\n%s\n```",
	// 	comment.User.Username,
	// 	comment.CreatedAt.Format("2006-01-02 15:04:05"),
	// 	comment.Content,
	// )

	htmlBody := `
		<html>
		<body>
			<h2>留言通知</h2>
			<p>作者：` + comment.User.Username + `</p>
			<p>時間：` + comment.CreatedAt.Format("2006-01-02 15:04:05") + `</p>
			<p>▼▼▼內容如下▼▼▼</p>
			<p>` + comment.Content + `</p>
			<p>網址：<a href="` + comment.URL + `">` + comment.URL + `</a></p>
			<br>
			<p>感謝您的留言！</p>
		</html>
		`

	email.SetBody(mail.TextHTML, htmlBody)

	return email.Send(smtpClient)
}

func getEnvAsInt(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}
