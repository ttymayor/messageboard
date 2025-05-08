# Message Board

看見 [go-restful-api-repository-messageboard](https://github.com/880831ian/go-restful-api-repository-messageboard) 該 Repo 後，發覺我的網站似乎沒有留言板，所以嘗試開發一個自己的留言板後端。

## Feature

- 獲得留言後，傳送 Email 通知（可選）

## 關於 Repo

### 如何使用

1. 確保已有 `Golang` 環境
2. 複製一份 `.env.example` 重新命名為 `.env` 並配置其中訊息
3. 啟動
    ```
    go run .
    ```

### 檔案結構

```
.
│   .env.example
│   .gitignore
│   go.mod
│   go.sum
│   main.go
│   README.md
│
├───controllers
│       auth.go
│       comment.go
│
├───middlewares
│       middleware.go
│
├───models
│       auth.go
│       model.go
│
└───routers
        router.go
```

### 使用套件

- Email Feature: github.com/xhit/go-simple-mail/v2
- 後端處理: github.com/gin-gonic/gin
- 環境變數: github.com/joho/godotenv
- PostgreSQL 連接: github.com/lib/pq

## To-Do

### 專案

- [ ] 支援 Docker 部屬

### 後端

目前只支援後端的 API 開發

依開發次序排序

- [x] ~~支援回覆功能~~ (Done)
- [x] ~~支援篩選每篇文章的留言~~ (Done)
- [x] ~~支援點讚~~ (Done)
- [x] ~~支援編輯留言~~ (Done)
- [ ] 支援隱藏留言
- [ ] 支援檢舉留言
- [ ] 支援不接收 Email 通知
- [ ] 支援 CAPTCHA 防刷機制
- [ ] 支援圖片或 Emoji 及 gif 等...
- [ ] 支援點讚 Email 通知，中介層防刷頻率

### 前端（可能另開 Repo）：

- [ ] 支援 Markdown
- [ ] 支援自訂配色
