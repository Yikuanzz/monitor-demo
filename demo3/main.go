package main

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	// 生产环境使用 NewProduction()
	logger, _ = zap.NewDevelopment() // 开发模式，输出更详细的信息
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 模拟业务逻辑
	time.Sleep(50 * time.Millisecond)

	// 结构化日志
	duration := time.Since(start)
	logger.Info("handled request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK),
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello with logs!"))
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	println("Server is running on port 9281")
	http.ListenAndServe(":9281", nil)
}
