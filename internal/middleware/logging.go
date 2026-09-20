package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

type LogContext struct {
	Action string
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	sw.body.Write(b)
	return sw.ResponseWriter.Write(b)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		logCtx := &LogContext{
			Action: fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		}

		sw := &statusWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(sw, r)

		userID := "guest"
		if uid, ok := r.Context().Value("user_id").(int); ok {
			userID = fmt.Sprintf("%d", uid)
		}

		result := "success"
		if sw.statusCode >= http.StatusBadRequest {
			result = fmt.Sprintf("(Error: %d - %s)", sw.statusCode, bytes.TrimSpace(sw.body.Bytes()))
		}

		log.Printf("[ %s | %s ] -> %d -> %v",
			t.Format("2006-01-02 15:04:05"),
			userID,
			logCtx.Action,
			result,
		)
	})
}
