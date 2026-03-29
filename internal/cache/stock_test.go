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

func TestResultValueCodec(t *testing.T) {
	value := BuildResultValue(ResultSuccess, 100)
	status, orderID, err := ParseResultValue(value)
	if err != nil {
		t.Fatalf("parse result value: %v", err)
	}
	if status != ResultSuccess || orderID != 100 {
		t.Fatalf("unexpected decode result: status=%s orderID=%d", status, orderID)
	}

	status, orderID, err = ParseResultValue(ResultFailed)
	if err != nil {
		t.Fatalf("parse failed status: %v", err)
	}
	if status != ResultFailed || orderID != 0 {
		t.Fatalf("unexpected decode result: status=%s orderID=%d", status, orderID)
	}
}

func TestNewStockCache(t *testing.T) {
	if _, err := NewStockCache(nil); err == nil {
		t.Fatal("expected nil redis error")
	}
}
