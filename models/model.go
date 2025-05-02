package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

type Comment struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Email     string `json:"email"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
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

	// 建立資料庫連線字串
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	// 連接資料庫
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("無法連接到資料庫：", err)
	}

	// 測試資料庫連線
	err = DB.Ping()
	if err != nil {
		log.Fatal("無法連接到資料庫：", err)
	}

	log.Println("成功連接到資料庫")

	// 建立留言資料表
	creatTable()
	log.Println("成功建立留言資料表")

	// // 建立索引
	// // 如果資料量很大，建議使用批次建立索引
	// _, err = DB.Exec(`CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments (created_at DESC);`)
	// if err != nil {
	// 	log.Fatal("無法建立索引：", err)
	// }
	// log.Println("成功建立索引")

}

func creatTable() {
	/*
	* 建立留言資料表
	* 表格名稱為 comments
	* id 欄位為主鍵，使用自動遞增的整數
	* author 欄位用來存放留言者的名稱
	* content 欄位用來存放留言內容
	* email 欄位用來存放留言者的電子郵件
	* created_at 欄位用來存放留言的時間戳
	 */

	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		author TEXT NOT NULL,
		email TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		log.Fatal("無法建立資料表：", err)
	}
}
