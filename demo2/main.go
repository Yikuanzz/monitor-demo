package main

import (
	"demo2/handler"
	"demo2/internal/metrics"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "demo2/internal/metrics" // 初始化指标注册
)

func main() {
	http.HandleFunc("/api/register", metrics.InstrumentHTTP("register", handler.RegisterUserHandler))
	http.Handle("/metrics", promhttp.Handler())
	println("Server is running on port 3055")
	http.ListenAndServe(":3055", nil)

}
