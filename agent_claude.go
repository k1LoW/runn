package runn

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Songmu/prompter"
	agent "github.com/k1LoW/claude-agent-sdk-go"
)

type claudeProvider struct {
	client *agent.Client
	opts   []agent.Option
}

func newClaudeProvider(cfg *AgentRunnerConfig) (*claudeProvider, error) {
	var opts []agent.Option

	if cfg.Model != "" {
		opts = append(opts, agent.WithModel(cfg.Model))
	}
	if cfg.System != "" {
		opts = append(opts, agent.WithSystemPrompt(cfg.System))
	}
	if len(cfg.Tools) > 0 {
		opts = append(opts, agent.WithTools(cfg.Tools...))
	}
	if cfg.Provider != "" && cfg.Provider != "anthropic" {
		return nil, fmt.Errorf("claude agent does not support provider %q (only anthropic)", cfg.Provider)
	}

	perms := parseAgentPermissions(cfg.Permissions)

	if perms.isAllowAll() {
		opts = append(opts, agent.WithPermissionMode("bypassPermissions"))
	} else {
		if len(perms.allowedTools) > 0 {
			opts = append(opts, agent.WithAllowedTools(perms.allowedTools...))
		}
	}
	if len(perms.deniedTools) > 0 {
		opts = append(opts, agent.WithDisallowedTools(perms.deniedTools...))
	}

	if perms.mode != "" {
		// Pass through as claude-specific permission mode (e.g., "plan", "acceptEdits")
		opts = append(opts, agent.WithPermissionMode(perms.mode))
	}
	// When no mode is set: no OnToolUse callback, so tool use requests will error
	// (safe default — agent cannot use tools without explicit permission)

	if cfg.Interactive {
		allowedTools := perms.allowedTools
		opts = append(opts, agent.WithOnToolUse(func(_ context.Context, toolName string, _ map[string]any, _ agent.ToolPermissionContext) (agent.PermissionResult, error) {
			if slices.Contains(allowedTools, toolName) {
				return &agent.PermissionAllow{}, nil
			}
			msg := fmt.Sprintf("Agent requests permission: %s", toolName)
			if prompter.YN(msg, false) {
				return &agent.PermissionAllow{}, nil
			}
			return &agent.PermissionDeny{Message: "denied by user"}, nil
		}))
		opts = append(opts, agent.WithOnAskUserQuestion(func(_ context.Context, q agent.Question) (string, error) {
			if len(q.Options) > 0 {
				labels := make([]string, len(q.Options))
				for i, o := range q.Options {
					labels[i] = o.Label
				}
				return prompter.Choose(q.Text, labels, labels[0]), nil
			}
			return prompter.Prompt(q.Text, ""), nil
		}))
	}

	return &claudeProvider{
		opts: opts,
	}, nil
}

func (p *claudeProvider) Run(ctx context.Context, req *agentRunRequest) (*AgentResponse, error) {
	if req.clearContext && p.client != nil {
		if err := p.client.Close(); err != nil {
			p.client = nil
			return nil, fmt.Errorf("claude agent close for context reset: %w", err)
		}
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
		p.closeAndReset()
		return nil, fmt.Errorf("claude agent send: %w", err)
	}

	var (
		buf    strings.Builder
		result string
	)
	for msg, err := range p.client.ReceiveResponse(ctx) {
		if err != nil {
			p.closeAndReset()
			return nil, fmt.Errorf("claude agent receive: %w", err)
		}
		switch m := msg.(type) {
		case *agent.AssistantMessage:
			for _, block := range m.Content {
				if tb, ok := block.(*agent.TextBlock); ok {
					buf.WriteString(tb.Text)
				}
			}
		case *agent.ResultMessage:
			if m.Result != "" {
				result = m.Result
			}
		}
	}

	content := result
	if content == "" {
		content = buf.String()
	}

	return &AgentResponse{Content: content}, nil
}

func (p *claudeProvider) Close() error {
	if p.client != nil {
		err := p.client.Close()
		p.client = nil
		return err
	}
	return nil
}

func (p *claudeProvider) closeAndReset() {
	if p.client != nil {
		_ = p.client.Close()
		p.client = nil
	}
}
