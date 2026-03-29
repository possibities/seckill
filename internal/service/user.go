package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"seckill/internal/model"
	"seckill/pkg"
)

type userReader interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type UserService struct {
	userRepo userReader
}

func NewUserService(userRepo userReader) (*UserService, error) {
	if userRepo == nil {
		return nil, pkg.ErrInternal
	}
	return &UserService{userRepo: userRepo}, nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (int64, error) {
	if username == "" || password == "" {
		return 0, pkg.ErrInvalidParam
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return 0, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, pkg.ErrUnauthorized
		}
		return 0, err
	}
	return user.ID, nil
}
