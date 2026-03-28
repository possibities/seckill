package pkg

import "testing"

func TestOK(t *testing.T) {
	resp := OK(map[string]any{"x": 1})
	if resp.Code != CodeOK {
		t.Fatalf("unexpected code: %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestFail(t *testing.T) {
	resp := Fail(ErrSoldOut)
	if resp.Code != CodeSoldOut {
		t.Fatalf("unexpected code: %d", resp.Code)
	}
	if resp.Message != "已售罄" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestFailNil(t *testing.T) {
	resp := Fail(nil)
	if resp.Code != CodeInternal {
		t.Fatalf("unexpected code: %d", resp.Code)
	}
}
