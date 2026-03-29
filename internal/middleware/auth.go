package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"seckill/config"
	"seckill/pkg"
)

const ContextUserIDKey = "user_id"

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
	expire time.Duration
}

func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret: []byte(cfg.Secret),
		issuer: cfg.Issuer,
		expire: cfg.Expire,
	}
}

func (m *JWTManager) Generate(userID int64) (string, int64, error) {
	now := time.Now()
	expireAt := now.Add(m.expire)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   "access_token",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}
	return signed, expireAt.Unix(), nil
}

func (m *JWTManager) Parse(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	if claims.Issuer != m.issuer {
		return 0, errors.New("invalid issuer")
	}
	return claims.UserID, nil
}

func (m *JWTManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.Fail(pkg.ErrUnauthorized))
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.Fail(pkg.ErrUnauthorized))
			return
		}

		uid, err := m.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.Fail(pkg.ErrUnauthorized))
			return
		}

		c.Set(ContextUserIDKey, uid)
		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}
