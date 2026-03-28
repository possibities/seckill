package model

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
	UserID   int64  `json:"user_id"`
}

type SeckillTokenResp struct {
	SeckillToken string `json:"seckill_token"`
	ExpireAt     int64  `json:"expire_at"`
}

type QueueResp struct {
	QueueStatus string `json:"queue_status"`
}

type SeckillResultResp struct {
	Status  string `json:"status"`
	OrderID int64  `json:"order_id,omitempty"`
}

type PayOrderResp struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	PaidAt  int64  `json:"paid_at"`
}
