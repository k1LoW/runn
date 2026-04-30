package runn

import (
	"context"
	"errors"
	"fmt"

	"github.com/Songmu/prompter"
	copilot "github.com/github/copilot-sdk/go"
)

type copilotProvider struct {
	client     *copilot.Client
	session    *copilot.Session
	clientOpts *copilot.ClientOptions
	sessionCfg *copilot.SessionConfig
}

func newCopilotProvider(cfg *AgentRunnerConfig) (*copilotProvider, error) {
	clientOpts := &copilot.ClientOptions{}

	sessionCfg := &copilot.SessionConfig{
		ClientName: "runn",
	}

	if cfg.Model != "" {
		sessionCfg.Model = cfg.Model
	}
	if cfg.System != "" {
		sessionCfg.SystemMessage = &copilot.SystemMessageConfig{
			Mode:    "custom",
			Content: cfg.System,
		}
	}
	if len(cfg.Tools) > 0 {
		sessionCfg.AvailableTools = cfg.Tools
	}
	if cfg.Provider != "" {
		sessionCfg.Provider = &copilot.ProviderConfig{
			Type: cfg.Provider,
		}
	}

	perms, err := parseAgentPermissions(cfg.Permissions)
	if err != nil {
		return nil, err
	}

	if perms.mode != "" {
		return nil, fmt.Errorf("unsupported copilot permissions value: %s", perms.mode)
	}
	if denied := perms.collectDenied(); len(denied) > 0 {
		sessionCfg.ExcludedTools = denied
	}

	interactive := cfg.Interactive
	sessionCfg.OnPermissionRequest = func(req copilot.PermissionRequest, _ copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
		toolName := ""
		if req.ToolName != nil {
			toolName = *req.ToolName
		}
		switch perms.decide(toolName) {
		case agentPermissionAllow:
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindApproved}, nil
		case agentPermissionDeny:
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindUserNotAvailable}, nil
		default:
			if interactive {
				msg := fmt.Sprintf("Agent requests permission: %s", toolName)
				if prompter.YN(msg, false) {
					return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindApproved}, nil
				}
				return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindRejected}, nil
			}
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindUserNotAvailable}, nil
		}
	}

	// Enable user input handling when interactive
	if cfg.Interactive {
		sessionCfg.OnUserInputRequest = func(req copilot.UserInputRequest, _ copilot.UserInputInvocation) (copilot.UserInputResponse, error) {
			answer := prompter.Prompt(req.Question, "")
			return copilot.UserInputResponse{Answer: answer, WasFreeform: true}, nil
		}
	}

	return &copilotProvider{
		clientOpts: clientOpts,
		sessionCfg: sessionCfg,
	}, nil
}

func (p *copilotProvider) Run(ctx context.Context, req *agentRunRequest) (*AgentResponse, error) {
	if req.clearContext && p.session != nil {
		if err := p.session.Disconnect(); err != nil {
			return nil, fmt.Errorf("copilot session disconnect: %w", err)
		}
		p.session = nil
	}

	if p.client == nil {
		p.client = copilot.NewClient(p.clientOpts)
		if err := p.client.Start(ctx); err != nil {
			p.client = nil
			return nil, fmt.Errorf("copilot client start: %w", err)
		}
	}

	if p.session == nil {
		session, err := p.client.CreateSession(ctx, p.sessionCfg)
		if err != nil {
			return nil, fmt.Errorf("copilot create session: %w", err)
		}
		p.session = session
	}

	event, err := p.session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: req.prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("copilot send: %w", err)
	}

	var content string
	if event != nil {
		if d, ok := event.Data.(*copilot.AssistantMessageData); ok {
			content = d.Content
		}
	}

	return &AgentResponse{Content: content}, nil
}

func (p *copilotProvider) Close() error {
	var errs error
	if p.session != nil {
		errs = errors.Join(errs, p.session.Disconnect())
		p.session = nil
	}
	if p.client != nil {
		errs = errors.Join(errs, p.client.Stop())
		p.client = nil
	}
	return errs
}
