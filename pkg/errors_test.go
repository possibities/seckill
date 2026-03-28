package pkg

import "testing"

func TestBizError(t *testing.T) {
	err := NewBizError(CodeInvalidParam, "参数错误")
	if err.Code != CodeInvalidParam {
		t.Fatalf("unexpected code: %d", err.Code)
	}
	if err.Message != "参数错误" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
	if err.Error() == "" {
		t.Fatal("error string should not be empty")
	}
}
