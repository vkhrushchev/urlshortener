package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
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

type loggedResponseWriter struct {
	http.ResponseWriter
	responseStatus int
	responseSize   int
}

func (lrw *loggedResponseWriter) Write(data []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(data)
	lrw.responseSize += size

	return size, err
}

func (lrw *loggedResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseStatus = statusCode
}

func LogRequestMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infow(
			"request_handling_started",
			"uri", r.RequestURI,
			"method", r.Method,
		)

		lrw := loggedResponseWriter{ResponseWriter: w}

		start := time.Now()
		next(&lrw, r)
		handlingTime := time.Since(start)

		log.Infow(
			"request_handling_ended",
			"handling_time_ms", handlingTime.Milliseconds(),
			"handling_time_ns", handlingTime.Nanoseconds(),
		)
		log.Infow(
			"response_data",
			"size", lrw.responseSize,
			"status", lrw.responseStatus)
	}
}

type compressResponseWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

func (crw *compressResponseWriter) Write(data []byte) (int, error) {
	contentType := crw.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		return crw.gzw.Write(data)
	}

	return crw.Write(data)
}

func (crw *compressResponseWriter) WriteHeader(statusCode int) {
	if statusCode >= 200 && statusCode < 300 {
		crw.Header().Set("Content-Encoding", "gzip")
	}

	crw.ResponseWriter.WriteHeader(statusCode)
}

func GzipMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tw := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			crw := &compressResponseWriter{
				ResponseWriter: w,
				gzw:            gzip.NewWriter(w),
			}
			tw = crw

			defer crw.gzw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Errorw(
					"decompress_error",
					"error", err.Error(),
				)

				tw.WriteHeader(http.StatusInternalServerError)
				tw.Write([]byte(err.Error()))

				return
			}

			r.Body = gzr
			defer r.Body.Close()
		}

		next(tw, r)
	}
}
