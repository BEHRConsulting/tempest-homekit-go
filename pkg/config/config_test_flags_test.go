package config

import (
	"flag"
	"os"
	"testing"
)

// TestTestEmailFlag tests the --test-email flag parsing
func TestTestEmailFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "valid email address",
			args:     []string{"-test-email", "user@example.com"},
			expected: "user@example.com",
		},
		{
			name:     "email with subdomain",
			args:     []string{"-test-email", "admin@mail.example.com"},
			expected: "admin@mail.example.com",
		},
		{
			name:     "email with plus sign",
			args:     []string{"-test-email", "user+test@example.com"},
			expected: "user+test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new FlagSet for each test
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			testEmail := fs.String("test-email", "", "Test email address")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if *testEmail != tt.expected {
				t.Errorf("Expected TestEmail=%s, got %s", tt.expected, *testEmail)
			}
		})
	}
}

// TestTestSMSFlag tests the --test-sms flag parsing
func TestTestSMSFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "valid US phone number",
			args:     []string{"-test-sms", "+15555551234"},
			expected: "+15555551234",
		},
		{
			name:     "valid international phone",
			args:     []string{"-test-sms", "+447911123456"},
			expected: "+447911123456",
		},
		{
			name:     "phone without plus sign",
			args:     []string{"-test-sms", "15555551234"},
			expected: "15555551234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			testSMS := fs.String("test-sms", "", "Test SMS phone number")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if *testSMS != tt.expected {
				t.Errorf("Expected TestSMS=%s, got %s", tt.expected, *testSMS)
			}
		})
	}
}

// TestTestConsoleFlagBool tests boolean test flags
func TestTestConsoleFlagBool(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		args     []string
		expected bool
	}{
		{
			name:     "test-console flag present",
			flagName: "test-console",
			args:     []string{"-test-console"},
			expected: true,
		},
		{
			name:     "test-syslog flag present",
			flagName: "test-syslog",
			args:     []string{"-test-syslog"},
			expected: true,
		},
		{
			name:     "test-oslog flag present",
			flagName: "test-oslog",
			args:     []string{"-test-oslog"},
			expected: true,
		},
		{
			name:     "test-eventlog flag present",
			flagName: "test-eventlog",
			args:     []string{"-test-eventlog"},
			expected: true,
		},
		{
			name:     "test-console flag absent",
			flagName: "test-console",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			testFlag := fs.Bool(tt.flagName, false, "Test flag")

			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if *testFlag != tt.expected {
				t.Errorf("Expected %s=%v, got %v", tt.flagName, tt.expected, *testFlag)
			}
		})
	}
}

// TestTestFlagsInConfig tests that test flags are properly loaded into Config struct
func TestTestFlagsInConfig(t *testing.T) {
	// Save and restore original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name            string
		envVars         map[string]string
		args            []string
		checkTestEmail  bool
		expectedEmail   string
		checkTestSMS    bool
		expectedSMS     string
		checkConsole    bool
		expectedConsole bool
		checkSyslog     bool
		expectedSyslog  bool
		checkOSLog      bool
		expectedOSLog   bool
		checkEventLog   bool
		expectedEventLog bool
	}{
		{
			name:           "test-email from command line",
			args:           []string{"cmd", "-test-email", "test@example.com"},
			checkTestEmail: true,
			expectedEmail:  "test@example.com",
		},
		{
			name:         "test-sms from command line",
			args:         []string{"cmd", "-test-sms", "+15555551234"},
			checkTestSMS: true,
			expectedSMS:  "+15555551234",
		},
		{
			name:            "test-console from command line",
			args:            []string{"cmd", "-test-console"},
			checkConsole:    true,
			expectedConsole: true,
		},
		{
			name:           "test-syslog from command line",
			args:           []string{"cmd", "-test-syslog"},
			checkSyslog:    true,
			expectedSyslog: true,
		},
		{
			name:          "test-oslog from command line",
			args:          []string{"cmd", "-test-oslog"},
			checkOSLog:    true,
			expectedOSLog: true,
		},
		{
			name:             "test-eventlog from command line",
			args:             []string{"cmd", "-test-eventlog"},
			checkEventLog:    true,
			expectedEventLog: true,
		},
		// Note: TestEmail and TestSMS do NOT support environment variables
		// They are command-line only flags
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
				defer os.Unsetenv(key)
			}

			// Set required environment variables to prevent validation errors
			os.Setenv("TEMPEST_TOKEN", "test-token")
			os.Setenv("TEMPEST_STATION_NAME", "TestStation")
			os.Setenv("HOMEKIT_PIN", "12345678")
			defer os.Unsetenv("TEMPEST_TOKEN")
			defer os.Unsetenv("TEMPEST_STATION_NAME")
			defer os.Unsetenv("HOMEKIT_PIN")

			// Set command line args
			os.Args = tt.args

			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg := LoadConfig()

			// Check expected values
			if tt.checkTestEmail && cfg.TestEmail != tt.expectedEmail {
				t.Errorf("Expected TestEmail=%s, got %s", tt.expectedEmail, cfg.TestEmail)
			}
			if tt.checkTestSMS && cfg.TestSMS != tt.expectedSMS {
				t.Errorf("Expected TestSMS=%s, got %s", tt.expectedSMS, cfg.TestSMS)
			}
			if tt.checkConsole && cfg.TestConsole != tt.expectedConsole {
				t.Errorf("Expected TestConsole=%v, got %v", tt.expectedConsole, cfg.TestConsole)
			}
			if tt.checkSyslog && cfg.TestSyslog != tt.expectedSyslog {
				t.Errorf("Expected TestSyslog=%v, got %v", tt.expectedSyslog, cfg.TestSyslog)
			}
			if tt.checkOSLog && cfg.TestOSLog != tt.expectedOSLog {
				t.Errorf("Expected TestOSLog=%v, got %v", tt.expectedOSLog, cfg.TestOSLog)
			}
			if tt.checkEventLog && cfg.TestEventLog != tt.expectedEventLog {
				t.Errorf("Expected TestEventLog=%v, got %v", tt.expectedEventLog, cfg.TestEventLog)
			}
		})
	}
}

// TestTestFlagsPrecedence tests that command-line flags override environment variables
func TestTestFlagsPrecedence(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set environment variables
	os.Setenv("TEST_EMAIL", "env@example.com")
	os.Setenv("TEST_SMS", "+10000000000")
	os.Setenv("TEMPEST_TOKEN", "test-token")
	os.Setenv("TEMPEST_STATION_NAME", "TestStation")
	os.Setenv("HOMEKIT_PIN", "12345678")
	defer os.Unsetenv("TEST_EMAIL")
	defer os.Unsetenv("TEST_SMS")
	defer os.Unsetenv("TEMPEST_TOKEN")
	defer os.Unsetenv("TEMPEST_STATION_NAME")
	defer os.Unsetenv("HOMEKIT_PIN")

	// Set command-line args that should override env vars
	os.Args = []string{
		"cmd",
		"-test-email", "cli@example.com",
		"-test-sms", "+15555551234",
	}

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg := LoadConfig()

	// Command-line should win
	if cfg.TestEmail != "cli@example.com" {
		t.Errorf("Expected CLI flag to override env var, got TestEmail=%s", cfg.TestEmail)
	}
	if cfg.TestSMS != "+15555551234" {
		t.Errorf("Expected CLI flag to override env var, got TestSMS=%s", cfg.TestSMS)
	}
}

// TestMultipleTestFlags tests behavior when multiple test flags are set
func TestMultipleTestFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TEMPEST_TOKEN", "test-token")
	os.Setenv("TEMPEST_STATION_NAME", "TestStation")
	os.Setenv("HOMEKIT_PIN", "12345678")
	defer os.Unsetenv("TEMPEST_TOKEN")
	defer os.Unsetenv("TEMPEST_STATION_NAME")
	defer os.Unsetenv("HOMEKIT_PIN")

	// Set multiple test flags at once
	os.Args = []string{
		"cmd",
		"-test-email", "test@example.com",
		"-test-sms", "+15555551234",
		"-test-console",
		"-test-syslog",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg := LoadConfig()

	// All flags should be set
	if cfg.TestEmail != "test@example.com" {
		t.Errorf("Expected TestEmail=test@example.com, got %s", cfg.TestEmail)
	}
	if cfg.TestSMS != "+15555551234" {
		t.Errorf("Expected TestSMS=+15555551234, got %s", cfg.TestSMS)
	}
	if !cfg.TestConsole {
		t.Error("Expected TestConsole=true")
	}
	if !cfg.TestSyslog {
		t.Error("Expected TestSyslog=true")
	}
}
