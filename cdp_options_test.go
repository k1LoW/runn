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
				// Base: DefaultExecAllocatorOptions + WSURLReadTimeout
				expectedOptCount := len(chromedp.DefaultExecAllocatorOptions) + 1
				// WindowSize is added only when flags don't contain a valid "window-size" string
				hasCustomWindowSize := false
				if v, ok := tt.flags["window-size"]; ok {
					if s, ok := v.(string); ok && s != "" {
						hasCustomWindowSize = true
					}
				}
				if !hasCustomWindowSize {
					expectedOptCount++ // +1 for default WindowSize
				}
				for range tt.flags {
					expectedOptCount++
				}

				if len(r.opts) < expectedOptCount {
					t.Errorf("opts not properly initialized, got %d options, expected at least %d", len(r.opts), expectedOptCount)
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
		{
			name: "invalid window-size type falls back to default",
			flags: map[string]any{
				"window-size": 12345,
			},
			wantCount: 1,
		},
		{
			name: "empty window-size falls back to default",
			flags: map[string]any{
				"window-size": "",
			},
			wantCount: 1,
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
			// Default options include DefaultExecAllocatorOptions + WSURLReadTimeout (+ WindowSize when no valid window-size flag)
			hasCustomWindowSize := false
			if v, ok := tt.flags["window-size"]; ok {
				if s, ok := v.(string); ok && s != "" {
					hasCustomWindowSize = true
				}
			}

			baseCount := len(chromedp.DefaultExecAllocatorOptions) + 1
			if !hasCustomWindowSize {
				baseCount++ // +1 for default WindowSize
			}

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

			// Verify no duplicate window-size: total opts should exactly match expected count
			if hasCustomWindowSize {
				expectedTotal := baseCount
				for range tt.flags {
					expectedTotal++
				}
				if len(r.opts) != expectedTotal {
					t.Errorf("expected exactly %d options with custom window-size (no duplicate), got %d", expectedTotal, len(r.opts))
				}
			}
		})
	}
}
