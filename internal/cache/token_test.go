package cache

import "testing"

func TestGenerateToken(t *testing.T) {
	a, err := GenerateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	b, err := GenerateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if len(a) != 32 || len(b) != 32 {
		t.Fatalf("unexpected token length: %d %d", len(a), len(b))
	}
	if a == b {
		t.Fatal("token should be random and unique")
	}
}

func TestNewTokenCache(t *testing.T) {
	if _, err := NewTokenCache(nil); err == nil {
		t.Fatal("expected nil redis error")
	}
}
