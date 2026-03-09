# 🚀 Buổi 1: Foundation & Project Setup

## Mục Tiêu

- ✅ Setup project structure theo Clean Architecture
- ✅ Hiểu Go basics và HTTP server
- ✅ Chạy được Hello World API

## Nội Dung Code

### 1. Project Structure

```
session1-foundation/
├── cmd/
│   └── server/
│       └── main.go          # Entry point - Hello World server
├── internal/
│   ├── model/               # (Empty - chuẩn bị buổi 2)
│   ├── handler/             # (Empty - chuẩn bị buổi 2)
│   ├── service/             # (Empty - chuẩn bị buổi 2)
│   └── storage/             # (Empty - chuẩn bị buổi 2)
├── go.mod
└── README.md
```

### 2. Key Concepts

#### HTTP Server trong Go

- `http.HandleFunc` - register handlers cho routes
- `http.ListenAndServe` - start server
- `ResponseWriter` và `Request` - handle HTTP

#### JSON Response

- `json.NewEncoder(w).Encode()` - convert Go struct sang JSON
- Content-Type header

### 3. Chạy Code

```bash
# Khởi tạo Go module
go mod init mini-asm

# Chạy server
go run cmd/server/main.go

# Test endpoint (terminal khác)
curl http://localhost:8080/health
```

**Expected Output:**

```json
{
  "status": "ok",
  "message": "Mini ASM service is running"
}
```

## So Sánh

### ❌ Bad Practice (Monolithic)

```go
// Tất cả code trong 1 file
func main() {
    http.HandleFunc("/assets", func(w http.ResponseWriter, r *http.Request) {
        // Parse request
        // Validate
        // Business logic
        // Database query
        // Response
    })
}
```

### ✅ Clean Architecture (Preview buổi 2)

```go
// Separation of concerns
handler → service → storage → model
```

**→ Buổi 1 setup structure, Buổi 2 implement layers!**

## Resources

- [Go HTTP Server Tutorial](https://gobyexample.com/http-servers)
- [Go Modules](https://go.dev/blog/using-go-modules)
- [REST API Best Practices](https://restfulapi.net/)
