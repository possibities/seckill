package repo

import (
	"testing"

	"gorm.io/gorm"
)

func TestNewGoodsRepo(t *testing.T) {
	if _, err := NewGoodsRepo(nil); err == nil {
		t.Fatal("expected error when db is nil")
	}

	r, err := NewGoodsRepo(&gorm.DB{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("repo should not be nil")
	}
}

func TestNewOrderRepo(t *testing.T) {
	if _, err := NewOrderRepo(nil); err == nil {
		t.Fatal("expected error when db is nil")
	}

	r, err := NewOrderRepo(&gorm.DB{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("repo should not be nil")
	}
}
