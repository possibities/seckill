package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"seckill/config"
)

func TestJWTGenerateAndParse(t *testing.T) {
	m := NewJWTManager(config.JWTConfig{Secret: "s", Issuer: "i", Expire: time.Hour})
	token, _, err := m.Generate(100)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	uid, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if uid != 100 {
		t.Fatalf("unexpected uid: %d", uid)
	}
}

func TestGlobalRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewGlobalRateLimiter(rate.Limit(1), 1))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected code: %d", w2.Code)
	}
}
