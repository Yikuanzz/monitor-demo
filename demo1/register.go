package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		userRegistrationsTotal,
		userRegistrationDuration,
		activeUsers,
	)
}

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
