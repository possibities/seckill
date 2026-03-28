package cache

import "testing"

func TestKeys(t *testing.T) {
	if got := StockKey(1); got != "sk:stock:1" {
		t.Fatalf("stock key mismatch: %s", got)
	}
	if got := SoldOutKey(2); got != "sk:sold_out:2" {
		t.Fatalf("sold out key mismatch: %s", got)
	}
	if got := UsersKey(3); got != "sk:users:3" {
		t.Fatalf("users key mismatch: %s", got)
	}
	if got := IdempotentTokenKey(4, 5); got != "sk:token:4:5" {
		t.Fatalf("token key mismatch: %s", got)
	}
	if got := ResultKey(6, 7); got != "sk:result:6:7" {
		t.Fatalf("result key mismatch: %s", got)
	}
	if got := URLTokenKey(8, 9); got != "sk:url_token:8:9" {
		t.Fatalf("url token key mismatch: %s", got)
	}
}
