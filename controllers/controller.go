package controllers

import (
	"net/http"
	"messageboard/models"
	"github.com/gin-gonic/gin"
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

