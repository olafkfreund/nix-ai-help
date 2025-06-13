package community

import (
	"testing"
)

func TestDiscourseClient_BasicInit_Current(t *testing.T) {
	client := NewDiscourseClient("https://discourse.nixos.org", "", "")
	if client == nil {
		t.Fatal("expected non-nil DiscourseClient")
	}
}
