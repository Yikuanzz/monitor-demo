package handler

import (
	"demo2/internal/metrics"
	"net/http"
	"time"
)

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 业务逻辑
	success := simulateUserRegistration()

	// 记录指标
	duration := time.Since(start).Seconds()
	metrics.UserRegistrationDuration.Observe(duration)

	if success {
		metrics.UserRegistrationsTotal.WithLabelValues("success").Inc()
		metrics.ActiveUsers.Inc()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User registered successfully"))
	} else {
		metrics.UserRegistrationsTotal.WithLabelValues("failed").Inc()
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User registration failed"))
	}
}

func simulateUserRegistration() bool {
	// 简单模拟：80% 成功
	return time.Now().Unix()%5 != 0
}
