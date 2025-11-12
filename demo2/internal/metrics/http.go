package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func InstrumentHTTP(handlerName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(ww, r)
		duration := time.Since(start).Seconds()
		HTTPRequestsTotal.WithLabelValues(r.Method, handlerName, strconv.Itoa(ww.statusCode)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, handlerName).Observe(duration)
	}
}
