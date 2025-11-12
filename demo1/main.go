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

var (
	userRegistrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations.",
		},
		[]string{"status"}, // "success" or "failed"
	)

	userRegistrationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_registration_duration_seconds",
			Help:    "Time spent processing user registration.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0},
		},
	)

	activeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Current number of active users.",
		},
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
	prometheus.MustRegister(
		httpRequestsTotal,        // HTTP 请求总数
		httpRequestDuration,      // HTTP 请求延迟
		userRegistrationsTotal,   // 用户注册总数
		userRegistrationDuration, // 用户注册延迟
		activeUsers,              // 活跃用户数
	)
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
