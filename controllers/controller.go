package controllers

import (
	"fmt"
	"log"
	"messageboard/models"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	mail "github.com/xhit/go-simple-mail/v2"
)

func CreateComment(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO comments (author, content) VALUES ($1, $2) RETURNING id, created_at`
	err := models.DB.QueryRow(query, comment.Author, comment.Content).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增留言失敗: " + err.Error()})
		return
	}

	// 成功新增後寄信通知
	if err := sendEmailNotification(comment); err != nil {
		// 不中斷流程，但可記錄錯誤
		fmt.Print(err)
	}
	log.Printf("成功寄送通知信給 %s\n", comment.Author)

	c.JSON(http.StatusOK, comment)
}

func GetComments(c *gin.Context) {
	rows, err := models.DB.Query(`SELECT id, author, content, created_at FROM comments ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢留言失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.Author, &comment.Content, &comment.CreatedAt)
		if err != nil {
			continue
		}
		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, comments)
}

func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	_, err := models.DB.Exec(`DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

func GetCommentByID(c *gin.Context) {
	id := c.Param("id")
	var comment models.Comment
	err := models.DB.QueryRow(`SELECT id, author, content, created_at FROM comments WHERE id = $1`, id).Scan(&comment.ID, &comment.Author, &comment.Content, &comment.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢留言失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
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
	email.SetFrom(os.Getenv("MAIL_FROM")).
		AddTo(os.Getenv("MAIL_TO")).
		SetSubject("【留言通知】你有一則新留言")

	createdAtStr := comment.CreatedAt
	createdAtFormatted := createdAtStr
	if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
		createdAtFormatted = t.Format("2006-01-02 15:04:05")
	}
	body := fmt.Sprintf("作者：%s\n內容：\n%s\n時間：%s", comment.Author, comment.Content, createdAtFormatted)
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
