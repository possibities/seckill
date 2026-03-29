package service

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"seckill/internal/model"
	"seckill/pkg"
)

type fakeUserRepo struct {
	user *model.User
	err  error
}

func (f *fakeUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

func TestAuthenticate(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	svc, _ := NewUserService(&fakeUserRepo{user: &model.User{ID: 1, PasswordHash: string(hash)}})
	uid, err := svc.Authenticate(context.Background(), "u", "pass")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if uid != 1 {
		t.Fatalf("unexpected uid: %d", uid)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	svc, _ := NewUserService(&fakeUserRepo{user: &model.User{ID: 1, PasswordHash: string(hash)}})
	_, err := svc.Authenticate(context.Background(), "u", "wrong")
	if err == nil || err.(*pkg.BizError).Code != pkg.CodeUnauthorized {
		t.Fatalf("expected unauthorized, got: %v", err)
	}
}
