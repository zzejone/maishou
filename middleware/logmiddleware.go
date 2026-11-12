// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (m *LogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 捕获请求体（请求体不会被外层缓冲，一定拿得到）
		var requestBody bytes.Buffer
		if r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			requestBody.Write(bodyBytes)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 尝试捕获响应体（注意：配置了 Timeout 时可能拿不到）
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next(recorder, r)

		fields := []logx.LogField{
			logx.Field("method", r.Method),
			logx.Field("path", r.URL.Path),
			logx.Field("query", r.URL.RawQuery),
			logx.Field("remote_addr", r.RemoteAddr),
			logx.Field("status", recorder.statusCode),
			logx.Field("duration_ms", time.Since(start).Milliseconds()),
			logx.Field("request_body", requestBody.String()),
		}

		// 只有捕获到响应内容才记录，避免空字段刷日志
		if recorder.body.Len() > 0 {
			fields = append(fields, logx.Field("response_body", recorder.body.String()))
		}

		logx.WithContext(r.Context()).Infow("HTTP Request Log", fields...)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
