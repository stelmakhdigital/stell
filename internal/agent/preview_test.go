package agent

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestToolResultPreview(t *testing.T) {
	tests := []struct {
		name, content, errText, want string
	}{
		{name: "empty", want: "ok"},
		{name: "short", content: "wrote 12 bytes to a.go", want: "wrote 12 bytes to a.go"},
		{name: "error wins", content: "ignored", errText: "pattern is required", want: "pattern is required"},
		{name: "collapse", content: "a.go\nb.go", want: "a.go b.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toolResultPreview(tt.content, tt.errText)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}

	long := strings.Repeat("abcdefghij\n", 20)
	got := toolResultPreview(long, "")
	if !strings.Contains(got, "lines") {
		t.Fatalf("expected line count, got %q", got)
	}
	if utf8.RuneCountInString(got) > previewMaxRunes+8 {
		t.Fatalf("too long: %q", got)
	}
	if strings.Contains(toolResultPreview("token=supersecret", ""), "supersecret") {
		t.Fatal("secret leaked")
	}
}
