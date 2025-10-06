package runn

import (
	"testing"
	"time"
)

func TestParseCDPRunnerWithTimeout(t *testing.T) {
	tests := []struct {
		name         string
		yamlConfig   string
		wantTimeout  time.Duration
		wantDetected bool
		wantErr      bool
	}{
		{
			name: "valid timeout configuration",
			yamlConfig: `
addr: chrome://new
timeout: 120sec
flags:
  headless: true
`,
			wantTimeout:  120 * time.Second,
			wantDetected: true,
			wantErr:      false,
		},
		{
			name: "valid timeout configuration cdp",
			yamlConfig: `
addr: cdp://new
timeout: 120sec
flags:
  headless: true
`,
			wantTimeout:  120 * time.Second,
			wantDetected: true,
			wantErr:      false,
		},
		{
			name: "timeout with minutes",
			yamlConfig: `
addr: chrome://new
timeout: 2m
`,
			wantTimeout:  2 * time.Minute,
			wantDetected: true,
			wantErr:      false,
		},
		{
			name: "timeout with complex duration",
			yamlConfig: `
addr: chrome://new
timeout: 1m30s
`,
			wantTimeout:  90 * time.Second,
			wantDetected: true,
			wantErr:      false,
		},
		{
			name: "no timeout specified - uses default",
			yamlConfig: `
addr: chrome://new
flags:
  headless: true
`,
			wantTimeout:  cdpTimeoutByStep, // default value
			wantDetected: true,
			wantErr:      false,
		},
		{
			name: "invalid timeout format",
			yamlConfig: `
addr: chrome://new
timeout: invalid
`,
			wantDetected: false,
			wantErr:      true,
		},
		{
			name:         "empty config - not detected as CDP runner",
			yamlConfig:   ``,
			wantDetected: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bk := &book{
				cdpRunners: map[string]*cdpRunner{},
			}

			detected, err := bk.parseCDPRunnerWithDetailed("test", []byte(tt.yamlConfig))

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("%v: parseCDPRunnerWithDetailed() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			// Check detection
			if detected != tt.wantDetected {
				t.Errorf("%v: parseCDPRunnerWithDetailed() detected = %v, want %v", tt.name, detected, tt.wantDetected)
				return
			}

			// If runner was created and no error, check timeout value
			if tt.wantDetected && !tt.wantErr {
				runner, ok := bk.cdpRunners["test"]
				if !ok {
					t.Error("CDP runner not found in book")
					return
				}
				if runner.timeoutByStep != tt.wantTimeout {
					t.Errorf("%v: timeout = %v, want %v", tt.name, runner.timeoutByStep, tt.wantTimeout)
				}
			}
		})
	}
}

func TestCDPTimeoutOption(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{
			name:    "valid timeout string",
			timeout: "30s",
		},
		{
			name:    "timeout in minutes",
			timeout: "5m",
		},
		{
			name:    "empty timeout - use default",
			timeout: "",
		},
		{
			name:    "invalid timeout format",
			timeout: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := CDPTimeout(tt.timeout)
			config := &cdpRunnerConfig{
				Flags:  map[string]any{},
				Remote: cdpNewKey,
			}

			err := opt(config)
			if err != nil {
				t.Errorf("CDPTimeout option returned error: %v", err)
				return
			}

			if config.Timeout != tt.timeout {
				t.Errorf("config.Timeout = %s, want %s", config.Timeout, tt.timeout)
			}
		})
	}
}

func TestApplyCDPTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     string
		wantTimeout time.Duration
		wantErr     bool
	}{
		{
			name:        "valid timeout",
			timeout:     "90sec",
			wantTimeout: 90 * time.Second,
			wantErr:     false,
		},
		{
			name:        "valid timeout min",
			timeout:     "3min",
			wantTimeout: 3 * time.Minute,
			wantErr:     false,
		},
		{
			name:        "valid timeout mix",
			timeout:     "3min30sec",
			wantTimeout: 3*time.Minute + 30*time.Second,
			wantErr:     false,
		},
		{
			name:        "empty timeout - no change",
			timeout:     "",
			wantTimeout: cdpTimeoutByStep, // default remains unchanged
			wantErr:     false,
		},
		{
			name:        "invalid timeout format",
			timeout:     "not-a-duration",
			wantTimeout: cdpTimeoutByStep,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock CDP runner
			runner := &cdpRunner{
				timeoutByStep: cdpTimeoutByStep, // start with default
			}

			err := applyCDPTimeout(runner, tt.timeout)

			if (err != nil) != tt.wantErr {
				t.Errorf("applyCDPTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && runner.timeoutByStep != tt.wantTimeout {
				t.Errorf("timeoutByStep = %v, want %v", runner.timeoutByStep, tt.wantTimeout)
			}
		})
	}
}
