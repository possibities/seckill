package model

import "testing"

func TestTableName(t *testing.T) {
	if (SeckillGoods{}).TableName() != "seckill_goods" {
		t.Fatalf("unexpected goods table name")
	}
	if (SeckillOrder{}).TableName() != "seckill_order" {
		t.Fatalf("unexpected order table name")
	}
	if (User{}).TableName() != "users" {
		t.Fatalf("unexpected user table name")
	}
}
