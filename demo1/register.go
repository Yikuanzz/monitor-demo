package main

import (
	"net/http"
	"time"
)

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 模拟注册逻辑
	success := simulateUserRegistration()

	duration := time.Since(start).Seconds()
	userRegistrationDuration.Observe(duration)

	if success {
		userRegistrationsTotal.WithLabelValues("success").Inc()
		activeUsers.Inc() // 假设注册即激活
		w.WriteHeader(http.StatusCreated)
	} else {
		userRegistrationsTotal.WithLabelValues("failed").Inc()
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Write([]byte("Registration processed"))
}

func simulateUserRegistration() bool {
	// 简单模拟：80% 成功
	return time.Now().Unix()%5 != 0
}
