package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ParentID  *uint     `json:"parent_id"`                          // 外鍵
	Parent    *Comment  `gorm:"foreignKey:ParentID" json:"parent"`  // 關聯
	Replies   []Comment `gorm:"foreignKey:ParentID" json:"replies"` // 子留言，一對多
	UserID    uint      `gorm:"not null" json:"user_id"`            // 外鍵
	User      User      `gorm:"foreignKey:UserID" json:"user"`      // 關聯
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"not null" json:"username"`
	Email     string    `gorm:"not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	RoleID    uint      `gorm:"not null" json:"role_id"`       // 外鍵
	Role      Role      `gorm:"foreignKey:RoleID" json:"role"` // 關聯
	LastLogin time.Time `json:"last_login"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoleName  string    `gorm:"not null" json:"role_name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func InitDB() {
	// 載入 .env 檔案
	err := godotenv.Load()
	if err != nil {
		log.Fatal("無法載入 .env 檔案：", err)
	}

	// 讀取環境變數
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// 建立 DSN 並連接資料庫
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("無法連接到資料庫：", err)
	}

	log.Println("成功連接到資料庫")

	// 僅開發環境下 drop table
	if os.Getenv("APP_ENV") == "dev" {
		DB.Migrator().DropTable(&Comment{}, &User{}, &Role{})
		log.Println("已刪除舊資料表")
	}

	// 自動建立資料表
	if err := DB.AutoMigrate(&User{}, &Role{}, &Comment{}); err != nil {
		log.Fatal("自動建立資料表失敗：", err)
	}
	log.Println("成功建立資料表")

	// 初始化預設角色
	InitRole()

	// 初始化預設使用者
	InitUser()
}

// 初始化身分
func InitRole() {
	// 建立預設角色
	var roles = []Role{
		{RoleName: "reader"},
		{RoleName: "admin"},
		{RoleName: "author"},
	}

	// 檢查角色是否存在，如果不存在則建立
	for _, role := range roles {
		var count int64
		DB.Model(&Role{}).Where("role_name = ?", role.RoleName).Count(&count)
		if count == 0 {
			if err := DB.Create(&role).Error; err != nil {
				log.Fatal("建立預設角色失敗：", err)
			}
			log.Printf("成功建立角色：%s\n", role.RoleName)
		} else {
			log.Printf("角色已存在：%s\n", role.RoleName)
		}
	}
}

// 初始化預設使用者
func InitUser() {
	// 建立預設使用者
	var users = []User{{
		Username: os.Getenv("AUTHOR_USERNAME"),
		Email:    os.Getenv("AUTHOR_EMAIL"),
		Password: os.Getenv("AUTHOR_PASSWORD"),
		RoleID:   3,
	}}

	for _, user := range users {
		var count int64
		DB.Model(&User{}).Where("email = ?", user.Email).Count(&count)
		if count == 0 {
			// 密碼加密
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Fatal("密碼加密失敗：", err)
			}
			user.Password = string(hashedPassword)

			if err := DB.Create(&user).Error; err != nil {
				log.Fatal("建立預設使用者失敗：", err)
			}
			log.Printf("成功建立使用者：%s\n", user.Username)
		} else {
			log.Printf("使用者已存在：%s\n", user.Username)
		}
	}
}
