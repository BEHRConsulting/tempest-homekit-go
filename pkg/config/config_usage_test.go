package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestCustomUsageIncludesAllFlags ensures that customUsage() prints the
// registered command-line flags so the help text doesn't drift from code.
func TestCustomUsageIncludesAllFlags(t *testing.T) {
	// Capture stderr written by customUsage()
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Call the usage printer
	customUsage()

	// Restore stderr and read buffer
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stderr = oldStderr
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read usage output: %v", err)
	}
	out := buf.String()

	// List of flags that must appear in the usage output. Keep this in sync
	// with flags registered in LoadConfig().
	expectedFlags := []string{
		"--token",
		"--station",
		"--pin",
		"--loglevel",
		"--logfilter",
		"--web-port",
		"--sensors",
		"--elevation",
		"--cleardb",
		"--disable-homekit",
		"--disable-alarms",
		"--history",
		"--history-read",
		"--history-reduce",
		"--history-reduce-method",
		"--history-bin-size",
		"--history-keep-recent-hours",
		"--chart-history",
		"--generate-path",
		"--alarms",
		"--alarms-edit",
		"--alarms-edit-port",
		"--webhook-listener",
		"--webhook-listener-port",
		"--env",
		"--status",
		"--status-refresh",
		"--status-timeout",
		"--status-theme",
		"--version",
		"--test-history",
		"--test-api",
		"--test-email",
		"--test-sms",
	}

	for _, flag := range expectedFlags {
		if !strings.Contains(out, flag) {
			t.Errorf("usage output missing expected flag: %s", flag)
		}
	}
}
