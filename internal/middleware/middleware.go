package middleware

import (
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/vkhrushchev/urlshortener/internal/common"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

type loggedResponseWriter struct {
	http.ResponseWriter
	responseStatus int
	responseSize   int
}

// Write переопределяет метод Write http.ResponseWriter
func (lrw *loggedResponseWriter) Write(data []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(data)
	lrw.responseSize += size

	return size, err
}

// WriteHeader переопределяет метод WriteHeader http.ResponseWriter
func (lrw *loggedResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseStatus = statusCode
}

// LogRequestMiddleware возвращает middleware для логирования запроса
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

// Header переопределяет метод Header http.ResponseWriter
func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

// Write переопределяет метод Write http.ResponseWriter
func (c *gzipWriter) Write(p []byte) (int, error) {
	// пишем через gzip.Writer для "text/html" и "application/json"
	contentType := c.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

// WriteHeader переопределяет метод WriteHeader http.ResponseWriter
func (c *gzipWriter) WriteHeader(statusCode int) {
	// для "text/html" и "application/json" проставляем заголовок "Content-Encoding: gzip"
	contentType := c.Header().Get("Content-Type")
	if contentType == "text/html" || contentType == "application/json" {
		c.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

// Close переопределяет метод Close http.ResponseWriter
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

// Read переопределяет метод Read io.ReadCloser
func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close переопределяет метод Close io.ReadCloser
func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware возвращает middleware для обработки запроса/ответа в формате gzip
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

// UserIDCookieMiddleware возвращает middleware для обработки кук "userID" и "userIDSignature"
func UserIDCookieMiddleware(salt string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var userIDCookie *http.Cookie
		var userIDSignatureCookie *http.Cookie
		var userID string
		var userIDSignature string
		var err error

		userIDCookie, err = r.Cookie("userID")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Errorw("middleware: error when get cookie 'userID'")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userIDSignatureCookie, err = r.Cookie("userIDSignature")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Errorw("middleware: error when get cookie 'userIDSignature'")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var isValidCookie bool
		if userIDCookie != nil {
			userID = userIDCookie.Value
			userIDSignature = userIDSignatureCookie.Value
			isValidCookie = common.CheckSignature(userID, userIDSignature, salt)
		}

		if userIDCookie == nil || !isValidCookie {
			log.Infow("middleware: 'userID' cookies not found or not valid")

			userID = uuid.NewString()
			userIDSignatureBytes := md5.Sum([]byte(userID + salt))
			userIDSignature = hex.EncodeToString(userIDSignatureBytes[:])

			userIDCookie = &http.Cookie{
				Name:   "userID",
				Value:  userID,
				Path:   "/",
				MaxAge: 3600,
			}

			userIDSignatureCookie = &http.Cookie{
				Name:   "userIDSignature",
				Value:  userIDSignature,
				Path:   "/",
				MaxAge: 3600,
			}

			http.SetCookie(w, userIDCookie)
			http.SetCookie(w, userIDSignatureCookie)
		}

		r = r.WithContext(context.WithValue(r.Context(), common.UserIDContextKey, userID))
		next(w, r)
	}
}

// AuthByUserIDCookieMiddleware возвращает middleware для авторизации по кукам "userID" и "userIDSignature"
func AuthByUserIDCookieMiddleware(salt string, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var userIDCookie *http.Cookie
		var userIDSignatureCookie *http.Cookie
		var err error

		userIDCookie, err = r.Cookie("userID")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Errorw("middleware: error when get cookie 'userID'")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userIDSignatureCookie, err = r.Cookie("userIDSignature")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Errorw("middleware: error when get cookie 'userIDSignature'")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var isValidCookie bool
		if userIDCookie != nil {
			isValidCookie = common.CheckSignature(userIDCookie.Value, userIDSignatureCookie.Value, salt)
		}

		if !isValidCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), common.UserIDContextKey, userIDCookie.Value))
		next(w, r)
	}
}

// CheckSubnetMiddleware возвращает middleware для проверки подсети из которой делается запрос
func CheckSubnetMiddleware(trustedSubnet *net.IPNet, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		realIP := r.Header.Get("X-Real-IP")

		if trustedSubnet == nil || !trustedSubnet.Contains(net.ParseIP(realIP)) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
