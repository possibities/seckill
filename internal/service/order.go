package service

import (
	"context"
	"fmt"
	"time"

	"seckill/internal/model"
	"seckill/pkg"
)

type orderAccessor interface {
	GetByID(ctx context.Context, orderID int64) (*model.SeckillOrder, error)
	UpdateStatusAndPaidAt(ctx context.Context, orderID int64, status int8, paidAt *time.Time) error
}

type OrderService struct {
	orderRepo orderAccessor
}

func NewOrderService(orderRepo orderAccessor) (*OrderService, error) {
	if orderRepo == nil {
		return nil, fmt.Errorf("nil order repo")
	}
	return &OrderService{orderRepo: orderRepo}, nil
}

func (s *OrderService) PayOrder(ctx context.Context, userID, orderID int64) (*model.PayOrderResp, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}
	if order.Status != model.OrderStatusPending {
		return nil, pkg.ErrInvalidParam
	}

	now := time.Now()
	if now.After(order.PayExpireAt) {
		_ = s.orderRepo.UpdateStatusAndPaidAt(ctx, orderID, model.OrderStatusCanceled, nil)
		return nil, pkg.ErrOrderTimeout
	}

	if err = s.orderRepo.UpdateStatusAndPaidAt(ctx, orderID, model.OrderStatusPaid, &now); err != nil {
		return nil, err
	}

	return &model.PayOrderResp{
		OrderID: orderID,
		Status:  "paid",
		PaidAt:  now.Unix(),
	}, nil
}
