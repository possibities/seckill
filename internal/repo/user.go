package repo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"seckill/internal/model"
	"seckill/pkg"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) (*UserRepo, error) {
	if db == nil {
		return nil, errors.New("nil gorm db")
	}
	return &UserRepo{db: db}, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUnauthorized
		}
		return nil, err
	}
	return &user, nil
}
