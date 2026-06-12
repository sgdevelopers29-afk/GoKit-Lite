package response

import "testing"

func TestSuccess(t *testing.T) {
	resp := Success("hello")

	if !resp.Success {
		t.Errorf("expected success=true")
	}

	if resp.Message != "success" {
		t.Errorf("expected success message")
	}

	if resp.Data != "hello" {
		t.Errorf("unexpected data")
	}
}

func TestError(t *testing.T) {
	resp := Error("not found")

	if resp.Success {
		t.Errorf("expected success=false")
	}

	if resp.Message != "not found" {
		t.Errorf("unexpected message")
	}
}
