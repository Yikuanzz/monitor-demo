package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconnds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// 定义中间件
func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建 ResponseWriter 包装器用于捕获状态码
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 执行实际请求 handler
		next(ww, r)

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.statusCode)
		method := r.Method
		endpoint := r.URL.Path

		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

func main() {
	// 使用中间件包装 handler
	http.HandleFunc("/api/hello", metricsMiddleware(helloHandler))
	http.HandleFunc("/api/error", metricsMiddleware(errorHandler))
	http.HandleFunc("/api/register", metricsMiddleware(registerUserHandler))
	http.Handle("/metrics", promhttp.Handler())
	println("Server is running on port 2365")
	http.ListenAndServe(":2365", nil)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(50 * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from Go!"))
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(150 * time.Millisecond)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}
