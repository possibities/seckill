package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"seckill/internal/model"
	"seckill/pkg"
)

type fakeOrderRepo struct {
	order         *model.SeckillOrder
	getErr        error
	updateErr     error
	updatedStatus int8
	updatedPaidAt *time.Time
}

func (f *fakeOrderRepo) GetByID(ctx context.Context, orderID int64) (*model.SeckillOrder, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.order, nil
}

func (f *fakeOrderRepo) UpdateStatusAndPaidAt(ctx context.Context, orderID int64, status int8, paidAt *time.Time) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updatedStatus = status
	f.updatedPaidAt = paidAt
	return nil
}

func TestPayOrderSuccess(t *testing.T) {
	repo := &fakeOrderRepo{order: &model.SeckillOrder{
		ID:          10,
		UserID:      1,
		Status:      model.OrderStatusPending,
		PayExpireAt: time.Now().Add(5 * time.Minute),
	}}
	svc, _ := NewOrderService(repo)

	resp, err := svc.PayOrder(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("pay order: %v", err)
	}
	if resp.Status != "paid" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if repo.updatedStatus != model.OrderStatusPaid {
		t.Fatalf("unexpected updated status: %d", repo.updatedStatus)
	}
	if repo.updatedPaidAt == nil {
		t.Fatal("paid_at should be set")
	}
}

func TestPayOrderTimeout(t *testing.T) {
	repo := &fakeOrderRepo{order: &model.SeckillOrder{
		ID:          11,
		UserID:      1,
		Status:      model.OrderStatusPending,
		PayExpireAt: time.Now().Add(-1 * time.Minute),
	}}
	svc, _ := NewOrderService(repo)

	_, err := svc.PayOrder(context.Background(), 1, 11)
	if !errors.Is(err, pkg.ErrOrderTimeout) {
		t.Fatalf("expected timeout error, got: %v", err)
	}
	if repo.updatedStatus != model.OrderStatusCanceled {
		t.Fatalf("expected canceled status, got: %d", repo.updatedStatus)
	}
}

func TestPayOrderOwnerMismatch(t *testing.T) {
	repo := &fakeOrderRepo{order: &model.SeckillOrder{
		ID:          12,
		UserID:      2,
		Status:      model.OrderStatusPending,
		PayExpireAt: time.Now().Add(5 * time.Minute),
	}}
	svc, _ := NewOrderService(repo)

	_, err := svc.PayOrder(context.Background(), 1, 12)
	if !errors.Is(err, pkg.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got: %v", err)
	}
}
