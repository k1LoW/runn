package runn

import (
	"context"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
)

func TestParseAgentRequest(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *AgentRequest
		wantErr bool
	}{
		{
			"simple prompt",
			`prompt: "Hello, world!"`,
			&AgentRequest{
				Prompt: "Hello, world!",
			},
			false,
		},
		{
			"prompt with clearContext",
			`
prompt: "New topic"
clearContext: true
`,
			&AgentRequest{
				Prompt:       "New topic",
				ClearContext: true,
			},
			false,
		},
		{
			"clearContext false",
			`
prompt: "Continue"
clearContext: false
`,
			&AgentRequest{
				Prompt:       "Continue",
				ClearContext: false,
			},
			false,
		},
		{
			"missing prompt",
			`clearContext: true`,
			nil,
			true,
		},
		{
			"empty map",
			`{}`,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v map[string]any
			if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
				t.Fatal(err)
			}
			got, err := parseAgentRequest(v)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error but got nil")
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestParseAgentRunnerConfig(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			"claude agent",
			`
agent: claude
model: sonnet
system: "You are a helpful assistant."
`,
			false,
		},
		{
			"copilot agent with provider",
			`
agent: copilot
provider: openai
model: gpt-5-nano
system: "You are a data analyst."
tools:
  - web_search
permissions:
  - "allow:*"
`,
			false,
		},
		{
			"missing agent field",
			`
model: sonnet
`,
			true,
		},
		{
			"missing model field",
			`
agent: claude
`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c AgentRunnerConfig
			if err := yaml.Unmarshal([]byte(tt.in), &c); err != nil {
				t.Fatal(err)
			}
			_, err := newAgentRunner(tt.name, &c)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error but got nil")
			}
		})
	}
}

func TestAgentRunnerDetection(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		detect bool
	}{
		{
			"agent runner detected",
			`
agent: claude
model: sonnet
`,
			true,
		},
		{
			"not agent runner - http",
			`
endpoint: https://example.com
`,
			false,
		},
		{
			"not agent runner - db",
			`
dsn: sqlite3://:memory:
`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected, err := newBook().parseAgentRunnerWithDetailed("test", []byte(tt.in))
			if err != nil {
				t.Fatal(err)
			}
			if detected != tt.detect {
				t.Errorf("got detected=%v, want %v", detected, tt.detect)
			}
		})
	}
}

type mockAgentProvider struct {
	response *AgentResponse
	err      error
	lastReq  *agentRunRequest
	closed   bool
}

func (m *mockAgentProvider) Run(_ context.Context, req *agentRunRequest) (*AgentResponse, error) {
	m.lastReq = req
	return m.response, m.err
}

func (m *mockAgentProvider) Close() error {
	m.closed = true
	return nil
}

func TestAgentRunnerRun(t *testing.T) {
	mock := &mockAgentProvider{
		response: &AgentResponse{Content: "Hello from agent"},
	}

	o, err := New(Book("testdata/book/always_success.yml"))
	if err != nil {
		t.Fatal(err)
	}

	rnr := &agentRunner{
		name:     "test-agent",
		provider: mock,
	}
	o.agentRunners["test-agent"] = rnr

	s := newStep(0, "test-step", o, map[string]any{})
	s.agentRunner = rnr
	s.agentRequest = map[string]any{
		"prompt": "Hello",
	}

	ctx := context.Background()
	if err := rnr.Run(ctx, s); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mock.lastReq == nil {
		t.Fatal("provider.Run was not called")
	}
	if mock.lastReq.prompt != "Hello" {
		t.Errorf("got prompt=%q, want %q", mock.lastReq.prompt, "Hello")
	}

	// Check store result
	latest := o.store.Latest()
	res, ok := latest["res"].(map[string]any)
	if !ok {
		t.Fatal("store does not contain res")
	}
	content, ok := res["content"].(string)
	if !ok {
		t.Fatal("store res does not contain content string")
	}
	if content != "Hello from agent" {
		t.Errorf("got content=%q, want %q", content, "Hello from agent")
	}
}

func TestNewClaudeProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *AgentRunnerConfig
		wantErr bool
	}{
		{
			"valid config",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet"},
			false,
		},
		{
			"with system and tools",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", System: "You are helpful.", Tools: []string{"Read", "Glob"}},
			false,
		},
		{
			"allow-all permissions",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"allow:*"}},
			false,
		},
		{
			"deny-all permissions",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"deny:*"}},
			false,
		},
		{
			"interactive permissions not supported",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"interactive"}},
			true,
		},
		{
			"SDK-specific permissions passthrough",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"plan"}},
			false,
		},
		{
			"allow and deny individual tools",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"allow:Read", "deny:Write"}},
			false,
		},
		{
			"mode with individual tool permissions",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Permissions: []string{"acceptEdits", "allow:Read", "allow:Bash"}},
			false,
		},
		{
			"invalid provider rejected",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Provider: "openai"},
			true,
		},
		{
			"anthropic provider accepted",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Provider: "anthropic"},
			false,
		},
		{
			"empty provider accepted",
			&AgentRunnerConfig{Agent: "claude", Model: "sonnet", Provider: ""},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newClaudeProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClaudeProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCopilotProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *AgentRunnerConfig
		wantErr bool
	}{
		{
			"valid config",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano"},
			false,
		},
		{
			"with provider and system",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Provider: "openai", System: "You are helpful."},
			false,
		},
		{
			"allow-all permissions",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"allow:*"}},
			false,
		},
		{
			"deny-all permissions",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"deny:*"}},
			false,
		},
		{
			"interactive permissions",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"interactive"}},
			false,
		},
		{
			"empty permissions defaults to deny",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano"},
			false,
		},
		{
			"unsupported permissions rejected",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"plan"}},
			true,
		},
		{
			"allow individual tools only",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"allow:Read", "allow:Bash"}},
			false,
		},
		{
			"interactive with allowed tools",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"interactive", "allow:Read"}},
			false,
		},
		{
			"deny individual tools",
			&AgentRunnerConfig{Agent: "copilot", Model: "gpt-5-nano", Permissions: []string{"allow:*", "deny:Write"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newCopilotProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCopilotProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentRunnerClose(t *testing.T) {
	mock := &mockAgentProvider{}
	rnr := &agentRunner{
		name:     "test",
		provider: mock,
	}
	if err := rnr.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !mock.closed {
		t.Error("provider.Close was not called")
	}
}
