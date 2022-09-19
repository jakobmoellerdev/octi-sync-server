package request_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"octi-sync-server/middleware/request"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestSizeLimiterOK(t *testing.T) {
	t.Parallel()
	router := gin.New()
	router.Use(request.BodySizeLimiter(10, gin.H{
		"msg": "request too large",
	}))
	router.POST("/test_ok", func(context *gin.Context) {
		_, _ = io.ReadAll(context.Request.Body)
		if len(context.Errors) > 0 {
			return
		}
		_ = context.Request.Body.Close()
		context.String(http.StatusOK, "OK")
	})
	resp := performRequest(http.MethodPost, "/test_ok", "big=abc", router)

	if resp.Code != http.StatusOK {
		t.Fatalf("error posting - http status %v", resp.Code)
	}
}

func TestRequestSizeLimiterOver(t *testing.T) {
	t.Parallel()
	router := gin.New()
	router.Use(request.BodySizeLimiter(10, gin.H{
		"msg": "request too large",
	}))
	router.POST("/test_large", func(context *gin.Context) {
		_, _ = io.ReadAll(context.Request.Body)
		if len(context.Errors) > 0 {
			return
		}
		_ = context.Request.Body.Close()
		context.String(http.StatusOK, "OK")
	})
	resp := performRequest(http.MethodPost, "/test_large", "big=abcdefghijklmnop", router)

	if resp.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("error posting - http status %v", resp.Code)
	}
}

func performRequest(method, target, body string, router http.Handler) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != "" {
		buf = new(bytes.Buffer)
		buf.WriteString(body)
	}
	r := httptest.NewRequest(method, target, buf)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}
