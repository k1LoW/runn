package runn

import (
	"context"
	"fmt"
	"strings"

	"github.com/Songmu/prompter"
	codex "github.com/k1LoW/codex-agent-sdk-go"
)

type codexProvider struct {
	client   *codex.Client
	threadID string
	opts     []codex.Option
	tOpts    []codex.ThreadOption
}

func newCodexProvider(cfg *AgentRunnerConfig) (*codexProvider, error) {
	if cfg.Provider != "" && cfg.Provider != "openai" {
		return nil, fmt.Errorf("codex agent does not support provider %q (only openai)", cfg.Provider)
	}

	perms := parseAgentPermissions(cfg.Permissions)

	var opts []codex.Option

	if cfg.Interactive {
		opts = append(opts, codex.WithOnCommandApproval(func(_ context.Context, req codex.CommandApprovalRequest) (codex.ApprovalDecision, error) {
			switch perms.decide(req.Command) {
			case agentPermissionAllow:
				return codex.DecisionAccept, nil
			case agentPermissionDeny:
				return codex.DecisionDecline, nil
			default:
				msg := fmt.Sprintf("Agent wants to run: %s", req.Command)
				if prompter.YN(msg, false) {
					return codex.DecisionAccept, nil
				}
				return codex.DecisionDecline, nil
			}
		}))
		opts = append(opts, codex.WithOnFileChangeApproval(func(_ context.Context, _ codex.FileChangeApprovalRequest) (codex.ApprovalDecision, error) {
			switch perms.decide("file_change") {
			case agentPermissionAllow:
				return codex.DecisionAccept, nil
			case agentPermissionDeny:
				return codex.DecisionDecline, nil
			default:
				if prompter.YN("Agent wants to modify files. Allow?", false) {
					return codex.DecisionAccept, nil
				}
				return codex.DecisionDecline, nil
			}
		}))
		opts = append(opts, codex.WithOnUserInput(func(_ context.Context, req codex.UserInputRequest) (map[string]string, error) {
			answers := make(map[string]string)
			for _, q := range req.Questions {
				id, _ := q["id"].(string)
				text, _ := q["text"].(string)
				answers[id] = prompter.Prompt(text, "")
			}
			return answers, nil
		}))
	} else {
		// Non-interactive: auto-approve or deny based on permissions
		opts = append(opts, codex.WithOnCommandApproval(func(_ context.Context, req codex.CommandApprovalRequest) (codex.ApprovalDecision, error) {
			if perms.decide(req.Command) == agentPermissionDeny {
				return codex.DecisionDecline, nil
			}
			return codex.DecisionAccept, nil
		}))
		opts = append(opts, codex.WithOnFileChangeApproval(func(_ context.Context, _ codex.FileChangeApprovalRequest) (codex.ApprovalDecision, error) {
			if perms.decide("file_change") == agentPermissionDeny {
				return codex.DecisionDecline, nil
			}
			return codex.DecisionAccept, nil
		}))
	}

	var tOpts []codex.ThreadOption
	if cfg.Model != "" {
		tOpts = append(tOpts, codex.WithModel(cfg.Model))
	}

	if perms.mode != "" {
		tOpts = append(tOpts, codex.WithApprovalPolicy(perms.mode))
	} else if perms.decide("*") == agentPermissionAllow {
		tOpts = append(tOpts, codex.WithApprovalPolicy("full-auto"))
	}
	if perms.sandbox != "" {
		tOpts = append(tOpts, codex.WithSandbox(perms.sandbox))
	}

	return &codexProvider{
		opts:  opts,
		tOpts: tOpts,
	}, nil
}

func (p *codexProvider) Run(ctx context.Context, req *agentRunRequest) (*AgentResponse, error) {
	if req.clearContext {
		if p.client != nil {
			_ = p.client.Close()
			p.client = nil
		}
		p.threadID = ""
	}

	if p.client == nil {
		c := codex.NewClient(p.opts...)
		if err := c.Connect(ctx); err != nil {
			return nil, fmt.Errorf("codex agent connect: %w", err)
		}
		p.client = c
	}

	if p.threadID == "" {
		thread, err := p.client.StartThread(ctx, p.tOpts...)
		if err != nil {
			p.closeAndReset()
			return nil, fmt.Errorf("codex start thread: %w", err)
		}
		p.threadID = thread.ID
	}

	var buf strings.Builder
	for evt, err := range p.client.StartTurn(ctx, p.threadID, []codex.UserInput{codex.TextInput(req.prompt)}) {
		if err != nil {
			p.closeAndReset()
			return nil, fmt.Errorf("codex turn: %w", err)
		}
		switch e := evt.(type) {
		case *codex.AgentMessageDeltaEvent:
			buf.WriteString(e.Delta)
		case *codex.ItemCompletedEvent:
			if msg, ok := e.Item.(*codex.AgentMessageItem); ok {
				// Use the complete text if available
				buf.Reset()
				buf.WriteString(msg.Text)
			}
		case *codex.ErrorEvent:
			p.closeAndReset()
			return nil, fmt.Errorf("codex error: %s", e.Message)
		}
	}

	return &AgentResponse{Content: buf.String()}, nil
}

func (p *codexProvider) Close() error {
	if p.client != nil {
		err := p.client.Close()
		p.client = nil
		p.threadID = ""
		return err
	}
	return nil
}

func (p *codexProvider) closeAndReset() {
	if p.client != nil {
		_ = p.client.Close()
		p.client = nil
		p.threadID = ""
	}
}
