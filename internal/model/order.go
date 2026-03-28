package model

import "time"

const (
	OrderStatusPending  int8 = 0
	OrderStatusPaid     int8 = 1
	OrderStatusCanceled int8 = 2
)

type SeckillOrder struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64      `gorm:"column:user_id;not null;index:uk_user_goods,unique" json:"user_id"`
	GoodsID       int64      `gorm:"column:goods_id;not null;index:uk_user_goods,unique" json:"goods_id"`
	SeckillPrice  float64    `gorm:"column:seckill_price;type:decimal(10,2);not null" json:"seckill_price"`
	Status        int8       `gorm:"column:status;not null" json:"status"`
	IdempotentKey string     `gorm:"column:idempotent_key;type:varchar(64);not null;uniqueIndex:uk_idempotent_key" json:"idempotent_key"`
	PayExpireAt   time.Time  `gorm:"column:pay_expire_at;not null;index:idx_status_expire" json:"pay_expire_at"`
	PaidAt        *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SeckillOrder) TableName() string {
	return "seckill_order"
}

type User struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	Username     string    `gorm:"column:username;type:varchar(64);not null;uniqueIndex:uk_username" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(128);not null" json:"password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
