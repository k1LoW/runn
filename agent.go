package runn

import (
	"context"
	"fmt"
	"sync"
)

const (
	agentStoreResponseKey = "res"
	agentStoreContentKey  = "content"
)

const (
	agentPermissionsAllowAll    = "allow-all"
	agentPermissionsDenyAll     = "deny-all"
	agentPermissionsInteractive = "interactive"
)

type agentRunner struct {
	name       string
	provider   agentProvider
	operatorID string
	mu         sync.Mutex
}

func newAgentRunner(name string, cfg *AgentRunnerConfig) (*agentRunner, error) {
	if cfg.Agent == "" {
		return nil, fmt.Errorf("agent runner %q requires agent field", name)
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("agent runner %q requires model field", name)
	}

	p, err := newAgentProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("agent runner %q: %w", name, err)
	}

	return &agentRunner{
		name:     name,
		provider: p,
	}, nil
}

func newAgentProvider(cfg *AgentRunnerConfig) (agentProvider, error) {
	switch cfg.Agent {
	case "claude":
		return newClaudeProvider(cfg)
	case "copilot":
		return newCopilotProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", cfg.Agent)
	}
}

func (rnr *agentRunner) Run(ctx context.Context, s *step) error {
	rnr.mu.Lock()
	defer rnr.mu.Unlock()

	o := s.parent
	e, err := o.expandBeforeRecord(s.agentRequest, s)
	if err != nil {
		return err
	}
	r, ok := e.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid agent request: %v", e)
	}
	parsed, err := parseAgentRequest(r)
	if err != nil {
		return err
	}

	o.capturers.captureAgentRequest(rnr.name, parsed)

	resp, err := rnr.provider.Run(ctx, &agentRunRequest{
		prompt:       parsed.Prompt,
		clearContext: parsed.ClearContext,
	})
	if err != nil {
		return err
	}

	o.capturers.captureAgentResponse(rnr.name, resp)

	o.record(s.idx, map[string]any{
		agentStoreResponseKey: map[string]any{
			agentStoreContentKey: resp.Content,
		},
	})
	return nil
}

func (rnr *agentRunner) Close() error {
	rnr.mu.Lock()
	defer rnr.mu.Unlock()

	if rnr.provider != nil {
		return rnr.provider.Close()
	}
	return nil
}
