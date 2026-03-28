package model

import "time"

const (
	GoodsStatusDraft     int8 = 0
	GoodsStatusPublished int8 = 1
	GoodsStatusRunning   int8 = 2
	GoodsStatusEnded     int8 = 3
)

type SeckillGoods struct {
	ID             int64     `gorm:"column:id;primaryKey" json:"id"`
	Name           string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	OriginalPrice  float64   `gorm:"column:original_price;type:decimal(10,2);not null" json:"original_price"`
	SeckillPrice   float64   `gorm:"column:seckill_price;type:decimal(10,2);not null" json:"seckill_price"`
	TotalStock     int       `gorm:"column:total_stock;not null" json:"total_stock"`
	AvailableStock int       `gorm:"column:available_stock;not null" json:"available_stock"`
	StartTime      time.Time `gorm:"column:start_time;not null" json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;not null" json:"end_time"`
	Status         int8      `gorm:"column:status;not null" json:"status"`
	ImgURL         string    `gorm:"column:img_url;type:varchar(512)" json:"img_url"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SeckillGoods) TableName() string {
	return "seckill_goods"
}
