package main

import (
	"os"

	"messageboard/models"
	"messageboard/routers"
)

func main() {
	// 初始化資料庫
	models.InitDB()

	// 設定環境模式與監聽位址
	env := os.Getenv("APP_ENV")
	var addr string
	if env == "dev" {
		addr = "127.0.0.1:8080"
	} else {
		addr = ":8080"
	}

	// 啟動服務
	// 註冊路由
	r := routers.SetupRouter()
	// 設定監聽的端口
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
