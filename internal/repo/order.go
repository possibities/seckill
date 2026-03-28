package repo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"seckill/internal/model"
	"seckill/pkg"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) (*OrderRepo, error) {
	if db == nil {
		return nil, errors.New("nil gorm db")
	}
	return &OrderRepo{db: db}, nil
}

func (r *OrderRepo) Create(ctx context.Context, order *model.SeckillOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepo) GetByID(ctx context.Context, orderID int64) (*model.SeckillOrder, error) {
	var order model.SeckillOrder
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) GetByUserAndGoods(ctx context.Context, userID, goodsID int64) (*model.SeckillOrder, error) {
	var order model.SeckillOrder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND goods_id = ?", userID, goodsID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) GetByIdempotentKey(ctx context.Context, key string) (*model.SeckillOrder, error) {
	var order model.SeckillOrder
	err := r.db.WithContext(ctx).Where("idempotent_key = ?", key).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) UpdateStatusAndPaidAt(ctx context.Context, orderID int64, status int8, paidAt *time.Time) error {
	updates := map[string]any{
		"status": status,
	}
	if paidAt != nil {
		updates["paid_at"] = *paidAt
	}

	result := r.db.WithContext(ctx).
		Model(&model.SeckillOrder{}).
		Where("id = ?", orderID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkg.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) ListExpiredPending(ctx context.Context, now time.Time, limit int) ([]model.SeckillOrder, error) {
	if limit <= 0 {
		limit = 100
	}
	var orders []model.SeckillOrder
	err := r.db.WithContext(ctx).
		Where("status = ? AND pay_expire_at <= ?", model.OrderStatusPending, now).
		Order("pay_expire_at ASC").
		Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
