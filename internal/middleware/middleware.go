package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	log = zap.Must(zap.NewProduction()).Sugar()
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
			"content-encoding", r.Header.Get("Content-Encoding"),
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

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	// пишем через gzip.Writer для "text/html" и "application/json"
	contentType := c.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	// для "text/html" и "application/json" проставляем заголовок "Content-Encoding: gzip"
	contentType := c.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		c.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	// закрываем только в том случае если писали через gzip.Writer, иначе записывается "GZIP footer" в конец ответа
	contentType := c.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		return c.zw.Close()
	}

	return nil
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newGzipWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next(ow, r)
	}
}
