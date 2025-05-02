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
│       controller.go
│
├───models
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

- 支援 Docker 部屬

### 後端

目前只支援後端的 API 開發

- 支援圖片或 Emoji 及 gif 等...
- 支援點讚與作者已讀紀錄
- 支援隱藏留言
- 支援回覆功能

### 前端（可能另開 Repo）：

- 支援 Markdown
- 支援自訂配色
