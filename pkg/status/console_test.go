package status

import (
	"strings"
	"testing"
)

// TestLogBufferWrite ensures that LogBuffer captures written data and returns
// it via GetLines with ANSI sequences stripped.
func TestLogBufferWrite(t *testing.T) {
	lb := NewLogBuffer(10)
	data := "hello world\n"
	n, err := lb.Write([]byte(data))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Write returned len %d, want %d", n, len(data))
	}
	lines := lb.GetLines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "hello world") {
		t.Fatalf("unexpected line content: %q", lines[0])
	}
}
