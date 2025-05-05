package controllers

import (
	"messageboard/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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
