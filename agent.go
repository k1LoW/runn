package runn

import (
	"context"
	"fmt"
)

const (
	agentStoreResponseKey = "res"
	agentStoreContentKey  = "content"
)

type agentRunner struct {
	name        string
	agent       string // agent type: "copilot", "claude", etc.
	provider    agentProvider
	model       string
	system      string
	tools       []string
	permissions string
	operatorID  string
}

func newAgentRunner(name string, cfg *agentRunnerConfig) (*agentRunner, error) {
	if cfg.Agent == "" {
		return nil, fmt.Errorf("agent runner %q requires agent field", name)
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("agent runner %q requires model field", name)
	}

	rnr := &agentRunner{
		name:        name,
		agent:       cfg.Agent,
		model:       cfg.Model,
		system:      cfg.System,
		tools:       cfg.Tools,
		permissions: cfg.Permissions,
	}

	p, err := newAgentProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("agent runner %q: %w", name, err)
	}
	rnr.provider = p

	return rnr, nil
}

func newAgentProvider(cfg *agentRunnerConfig) (agentProvider, error) {
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
	if rnr.provider != nil {
		return rnr.provider.Close()
	}
	return nil
}
