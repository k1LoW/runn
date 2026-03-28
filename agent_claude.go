package runn

import (
	"context"
	"fmt"

	agent "github.com/k1LoW/claude-agent-sdk-go"
)

type claudeProvider struct {
	client *agent.Client
	opts   []agent.Option
}

func newClaudeProvider(cfg *agentRunnerConfig) (*claudeProvider, error) {
	var opts []agent.Option

	if cfg.Model != "" {
		opts = append(opts, agent.WithModel(cfg.Model))
	}
	if cfg.System != "" {
		opts = append(opts, agent.WithSystemPrompt(cfg.System))
	}
	if len(cfg.Tools) > 0 {
		opts = append(opts, agent.WithAllowedTools(cfg.Tools...))
	}

	// Map permissions to claude-specific settings
	switch cfg.Permissions {
	case "allow-all":
		opts = append(opts, agent.WithPermissionMode("bypassPermissions"))
	case "deny-all":
		opts = append(opts, agent.WithDisallowedTools("*"))
	case "interactive":
		// Will be handled at Run time with OnToolUse/OnAskUserQuestion callbacks
	case "":
		// Default: no special permission mode
	default:
		// Pass through as claude-specific permission mode (e.g., "plan", "acceptEdits")
		opts = append(opts, agent.WithPermissionMode(cfg.Permissions))
	}

	return &claudeProvider{
		opts: opts,
	}, nil
}

func (p *claudeProvider) Run(ctx context.Context, req *agentRunRequest) (*AgentResponse, error) {
	if req.clearContext && p.client != nil {
		_ = p.client.Close()
		p.client = nil
	}

	if p.client == nil {
		c := agent.NewClient(p.opts...)
		if err := c.Connect(ctx); err != nil {
			return nil, fmt.Errorf("claude agent connect: %w", err)
		}
		p.client = c
	}

	if err := p.client.Send(ctx, req.prompt); err != nil {
		return nil, fmt.Errorf("claude agent send: %w", err)
	}

	var content string
	for msg, err := range p.client.ReceiveResponse(ctx) {
		if err != nil {
			return nil, fmt.Errorf("claude agent receive: %w", err)
		}
		switch m := msg.(type) {
		case *agent.AssistantMessage:
			for _, block := range m.Content {
				if tb, ok := block.(*agent.TextBlock); ok {
					content += tb.Text
				}
			}
		case *agent.ResultMessage:
			if m.Result != "" {
				content = m.Result
			}
		}
	}

	return &AgentResponse{Content: content}, nil
}

func (p *claudeProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
