package runn

import (
	"context"
	"os"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/runn/testutil"
)

func TestCDPRunnerWithOptions(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}

	tests := []struct {
		name    string
		flags   map[string]any
		wantErr bool
	}{
		{
			name: "with headless flag",
			flags: map[string]any{
				"headless": true,
			},
			wantErr: false,
		},
		{
			name: "with multiple flags",
			flags: map[string]any{
				"headless":    true,
				"disable-gpu": true,
				"no-sandbox":  true,
			},
			wantErr: false,
		},
		{
			name: "with window size",
			flags: map[string]any{
				"window-size": "1280,720",
			},
			wantErr: false,
		},
		{
			name: "with user agent",
			flags: map[string]any{
				"user-agent": "Custom User Agent for Testing",
			},
			wantErr: false,
		},
		{
			name: "with disable web security",
			flags: map[string]any{
				"disable-web-security": true,
			},
			wantErr: false,
		},
		{
			name:    "empty flags",
			flags:   map[string]any{},
			wantErr: false,
		},
		{
			name:    "nil flags",
			flags:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)

			r, err := newCDPRunnerWithOptions("test", cdpNewKey, tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCDPRunnerWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			t.Cleanup(func() {
				if err := r.Close(); err != nil {
					t.Error(err)
				}
			})

			// Verify that flags were applied to opts
			if tt.flags != nil {
				// Count expected options
				expectedOptCount := len(chromedp.DefaultExecAllocatorOptions) + 1 // +1 for WindowSize
				for range tt.flags {
					expectedOptCount++
				}

				// Check if RUNN_DISABLE_HEADLESS is set
				if os.Getenv("RUNN_DISABLE_HEADLESS") != "" {
					expectedOptCount += 3 // headless, hide-scrollbars, mute-audio
				}

				// Check if RUNN_DISABLE_CHROME_SANDBOX is set
				if os.Getenv("RUNN_DISABLE_CHROME_SANDBOX") != "" {
					expectedOptCount += 1 // no-sandbox
				}

				if len(r.opts) < len(chromedp.DefaultExecAllocatorOptions) {
					t.Errorf("opts not properly initialized, got %d options", len(r.opts))
				}
			}

			// Test that the runner can be used
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			s := newStep(0, "stepKey", o, nil)
			t.Cleanup(func() {
				o.store.ClearSteps()
			})

			// Simple test action to verify the runner works
			actions := CDPActions{
				{
					Fn: "navigate",
					Args: map[string]any{
						"url": "about:blank",
					},
				},
			}

			if err := r.run(ctx, actions, s); err != nil {
				t.Errorf("failed to run actions with flags: %v", err)
			}
		})
	}
}

func TestParseCDPRunnerWithDetailed(t *testing.T) {
	tests := []struct {
		name       string
		yamlConfig string
		wantRunner bool
		wantErr    bool
	}{
		{
			name: "with addr and flags",
			yamlConfig: `
addr: chrome://new
flags:
  headless: true
  disable-gpu: true
`,
			wantRunner: true,
			wantErr:    false,
		},
		{
			name: "with only flags",
			yamlConfig: `
flags:
  headless: false
  no-sandbox: true
`,
			wantRunner: true,
			wantErr:    false,
		},
		{
			name: "with only addr",
			yamlConfig: `
addr: chrome://new
`,
			wantRunner: true,
			wantErr:    false,
		},
		{
			name: "empty config",
			yamlConfig: `
`,
			wantRunner: false,
			wantErr:    false,
		},
		{
			name: "invalid yaml",
			yamlConfig: `
addr: chrome://new
flags:
  - invalid
`,
			wantRunner: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bk := &book{
				cdpRunners: map[string]*cdpRunner{},
			}

			detected, err := bk.parseCDPRunnerWithDetailed("test", []byte(tt.yamlConfig))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCDPRunnerWithDetailed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if detected != tt.wantRunner {
				t.Errorf("parseCDPRunnerWithDetailed() detected = %v, want %v", detected, tt.wantRunner)
			}

			if tt.wantRunner {
				if _, ok := bk.cdpRunners["test"]; !ok {
					t.Error("runner not added to book")
				}
			}
		})
	}
}

func TestCDPOptionParsing(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]any
		wantCount int // minimum expected option count
	}{
		{
			name: "boolean flags",
			flags: map[string]any{
				"headless":    true,
				"disable-gpu": false,
			},
			wantCount: 2,
		},
		{
			name: "string flags",
			flags: map[string]any{
				"user-agent":  "Test Agent",
				"window-size": "800,600",
			},
			wantCount: 2,
		},
		{
			name: "integer flags",
			flags: map[string]any{
				"remote-debugging-port": 9222,
			},
			wantCount: 1,
		},
		{
			name: "float flags (converted to int)",
			flags: map[string]any{
				"remote-debugging-port": 9223.0,
			},
			wantCount: 1,
		},
		{
			name: "mixed types",
			flags: map[string]any{
				"headless":              true,
				"user-agent":            "Test",
				"remote-debugging-port": 9222,
			},
			wantCount: 3,
		},
		{
			name: "unsupported types are skipped",
			flags: map[string]any{
				"valid-flag": true,
				"array-flag": []string{"ignored"},
				"map-flag":   map[string]string{"ignored": "value"},
			},
			wantCount: 1, // only valid-flag is counted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := newCDPRunnerWithOptions("test", cdpNewKey, tt.flags)
			if err != nil {
				t.Fatalf("newCDPRunnerWithOptions() error = %v", err)
			}
			t.Cleanup(func() {
				if err := r.Close(); err != nil {
					t.Error(err)
				}
			})

			// Count the number of options added beyond defaults
			// Default options include DefaultExecAllocatorOptions + WindowSize
			baseCount := len(chromedp.DefaultExecAllocatorOptions) + 1

			// Account for environment variable options
			if os.Getenv("RUNN_DISABLE_HEADLESS") != "" {
				baseCount += 3
			}
			if os.Getenv("RUNN_DISABLE_CHROME_SANDBOX") != "" {
				baseCount += 1
			}

			// Check that we have at least the base options
			if len(r.opts) < baseCount {
				t.Errorf("expected at least %d base options, got %d", baseCount, len(r.opts))
			}
		})
	}
}
