package cache

import "testing"

func TestParseScriptResult(t *testing.T) {
	res, err := parseScriptResult(int64(1))
	if err != nil || res != StockSuccess {
		t.Fatalf("unexpected parse result: %v %v", res, err)
	}

	res, err = parseScriptResult("-1")
	if err != nil || res != StockNoInventory {
		t.Fatalf("unexpected parse result: %v %v", res, err)
	}

	if _, err = parseScriptResult("bad"); err == nil {
		t.Fatal("expected parse error for bad string")
	}

	if _, err = parseScriptResult(struct{}{}); err == nil {
		t.Fatal("expected parse error for unknown type")
	}
}

func TestNewStockCache(t *testing.T) {
	if _, err := NewStockCache(nil); err == nil {
		t.Fatal("expected nil redis error")
	}
}
