# Message Board

看見 [go-restful-api-repository-messageboard](https://github.com/880831ian/go-restful-api-repository-messageboard) 該 Repo 後，發覺我的網站似乎沒有留言板，所以嘗試開發一個自己的留言板後端。

## Feature

- 可設置來源許可，防止 CSRF
- 獲得留言後，傳送 Email 通知（可選）
- 支援 Docker 部署

## 關於 Repo

完成進度：`70%`
請注意：`.env` 檔內若設置 `APP_ENV=dev` 每次執行專案時，會刪表重建資料庫。
欲保留數據，請更改為 `APP_ENV=prod`

### 使用方法一：直接運行

1. 確保已有 `Golang` 環境
2. 複製一份 `.env.example` 重新命名為 `.env` 並配置其中訊息
3. 啟動
   ```
   go run .
   ```

### 使用方法二：使用 Docker

1. 確保已安裝 `Docker` 和 `Docker Compose`
2. 複製一份 `.env.example` 重新命名為 `.env` 並配置其中訊息
3. 使用 Docker Compose 啟動應用
   ```
   docker-compose up -d
   ```
4. 應用將在 `http://localhost:8080` 運行

### 使用套件

- CORS 跨站處理: github.com/gin-contrib/cors
- 登入驗證機制: github.com/golang-jwt/jwt/v5
- Email Feature: github.com/xhit/go-simple-mail/v2
- 後端處理: github.com/gin-gonic/gin
- 環境變數: github.com/joho/godotenv
- 資料庫操作: gorm.io/gorm
- PostgreSQL 連接: gorm.io/driver/postgres

## To-Do

### 專案

- [x] 支援 Docker 部屬

### 後端

目前只支援後端的 API 開發

依開發次序排序

- [x] ~~支援回覆功能~~ (Done)
- [x] ~~支援篩選每篇文章的留言~~ (Done)
- [x] ~~支援點讚~~ (Done)
- [x] ~~支援編輯留言~~ (Done)
- [x] ~~防 CSRF，使用 github.com/gin-contrib/cors 套件防跨站請求~~ (Done)
- [x] ~~限制速率，使用 golang.org/x/time/rate 中介層~~ (Done)
- [ ] 支援隱藏留言
- [ ] 支援檢舉留言
- [ ] 支援不接收 Email 通知
- [ ] 支援 CAPTCHA 防刷機制
- [ ] 支援圖片或 Emoji 及 gif 等...
- [ ] 支援點讚 Email 通知，中介層防刷頻率

### 前端（可能另開 Repo）：

- [ ] 支援 Markdown
- [ ] 支援自訂配色
