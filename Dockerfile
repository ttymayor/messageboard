# 使用官方 Golang 鏡像作為構建階段
FROM golang:1.24-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -o messageboard .

# 使用輕量級的 alpine 鏡像作為運行階段
FROM alpine:latest

# 安裝 ca-certificates 以支持 HTTPS
RUN apk --no-cache add ca-certificates

# 設置工作目錄
WORKDIR /root/

# 從構建階段複製編譯好的二進制文件
COPY --from=builder /app/messageboard .

# 複製環境變數配置文件
COPY .env .

# 暴露應用端口
EXPOSE 8080

# 運行應用
CMD ["./messageboard"]