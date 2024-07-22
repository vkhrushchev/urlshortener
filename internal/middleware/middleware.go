package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic("cannot initilize zap")
	}

	log = zapLogger.Sugar()
}

type extendedResponseWriter struct {
	http.ResponseWriter
	responseStatus int
	responseSize   int
}

func (erw *extendedResponseWriter) Write(data []byte) (int, error) {
	size, err := erw.ResponseWriter.Write(data)
	erw.responseSize += size

	return size, err
}

func (erw *extendedResponseWriter) WriteHeader(statusCode int) {
	erw.ResponseWriter.WriteHeader(statusCode)
	erw.responseStatus = statusCode
}

func LogRequest(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infow(
			"request_handling_started",
			"uri", r.RequestURI,
			"method", r.Method,
		)

		erw := extendedResponseWriter{ResponseWriter: w}

		start := time.Now()
		next(&erw, r)
		handlingTime := time.Since(start)

		log.Infow(
			"request_handling_ended",
			"handling_time_ms", handlingTime.Milliseconds(),
			"handling_time_ns", handlingTime.Nanoseconds(),
		)
		log.Infow(
			"response_data",
			"size", erw.responseSize,
			"status", erw.responseStatus)

		fmt.Println("<-- request handling")
	}
}
