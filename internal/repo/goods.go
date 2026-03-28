package repo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"seckill/internal/model"
	"seckill/pkg"
)

type GoodsRepo struct {
	db *gorm.DB
}

func NewGoodsRepo(db *gorm.DB) (*GoodsRepo, error) {
	if db == nil {
		return nil, errors.New("nil gorm db")
	}
	return &GoodsRepo{db: db}, nil
}

func (r *GoodsRepo) GetByID(ctx context.Context, goodsID int64) (*model.SeckillGoods, error) {
	var goods model.SeckillGoods
	err := r.db.WithContext(ctx).Where("id = ?", goodsID).First(&goods).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrGoodsNotFound
		}
		return nil, err
	}
	return &goods, nil
}

func (r *GoodsRepo) DecreaseAvailableStock(ctx context.Context, goodsID int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.SeckillGoods{}).
		Where("id = ? AND available_stock > 0", goodsID).
		UpdateColumn("available_stock", gorm.Expr("available_stock - 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkg.ErrSoldOut
	}
	return nil
}

func (r *GoodsRepo) UpdateStatus(ctx context.Context, goodsID int64, status int8) error {
	result := r.db.WithContext(ctx).
		Model(&model.SeckillGoods{}).
		Where("id = ?", goodsID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkg.ErrGoodsNotFound
	}
	return nil
}
