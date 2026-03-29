package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"seckill/internal/middleware"
	"seckill/internal/model"
	"seckill/pkg"
)

type userAuthenticator interface {
	Authenticate(ctx context.Context, username, password string) (int64, error)
}

type UserHandler struct {
	auth userAuthenticator
	jwt  *middleware.JWTManager
}

func NewUserHandler(auth userAuthenticator, jwt *middleware.JWTManager) (*UserHandler, error) {
	if auth == nil || jwt == nil {
		return nil, pkg.ErrInternal
	}
	return &UserHandler{auth: auth, jwt: jwt}, nil
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, pkg.ErrInvalidParam)
		return
	}

	userID, err := h.auth.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}

	token, expireAt, err := h.jwt.Generate(userID)
	if err != nil {
		writeError(c, err)
		return
	}

	writeOK(c, model.LoginResp{
		Token:    token,
		ExpireAt: expireAt,
		UserID:   userID,
	})
}
