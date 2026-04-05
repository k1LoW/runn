package runn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/k1LoW/runn/internal/scope"
)

const (
	agentStoreResponseKey = "res"
	agentStoreContentKey  = "content"
)

const (
	agentPermissionsAllowPrefix   = "allow:"
	agentPermissionsDenyPrefix    = "deny:"
	agentPermissionsSandboxPrefix = "sandbox:"
)

type agentPermissionDecision int

const (
	agentPermissionUndecided agentPermissionDecision = iota
	agentPermissionAllow
	agentPermissionDeny
)

// agentParsedPermissions holds the permissions rules and SDK-specific settings.
// Rules are evaluated in order (last match wins).
type agentParsedPermissions struct {
	rules   []agentPermissionRule // ordered rules from permissions array
	mode    string               // SDK-specific mode (e.g., "plan", "acceptEdits", "full-auto")
	sandbox string               // sandbox mode (e.g., "workspace-write", "workspace-read")
}

type agentPermissionRule struct {
	prefix   string // "allow" or "deny"
	toolName string // tool name or "*" for wildcard
}

func parseAgentPermissions(perms []string) (*agentParsedPermissions, error) {
	p := &agentParsedPermissions{}
	for _, perm := range perms {
		perm = strings.TrimSpace(perm)
		if perm == "" {
			continue
		}
		switch {
		case strings.HasPrefix(perm, agentPermissionsAllowPrefix):
			toolName := strings.TrimPrefix(perm, agentPermissionsAllowPrefix)
			if toolName == "" {
				return nil, fmt.Errorf("invalid permission rule %q: tool name is required", perm)
			}
			p.rules = append(p.rules, agentPermissionRule{
				prefix:   "allow",
				toolName: toolName,
			})
		case strings.HasPrefix(perm, agentPermissionsDenyPrefix):
			toolName := strings.TrimPrefix(perm, agentPermissionsDenyPrefix)
			if toolName == "" {
				return nil, fmt.Errorf("invalid permission rule %q: tool name is required", perm)
			}
			p.rules = append(p.rules, agentPermissionRule{
				prefix:   "deny",
				toolName: toolName,
			})
		case strings.HasPrefix(perm, agentPermissionsSandboxPrefix):
			sandbox := strings.TrimPrefix(perm, agentPermissionsSandboxPrefix)
			if sandbox == "" {
				return nil, fmt.Errorf("invalid permission rule %q: sandbox mode is required", perm)
			}
			p.sandbox = sandbox
		default:
			p.mode = perm
		}
	}
	return p, nil
}

// decide evaluates the permission rules for a tool (last match wins).
func (p *agentParsedPermissions) decide(toolName string) agentPermissionDecision {
	result := agentPermissionUndecided
	for _, rule := range p.rules {
		if rule.toolName == "*" || rule.toolName == toolName {
			switch rule.prefix {
			case "allow":
				result = agentPermissionAllow
			case "deny":
				result = agentPermissionDeny
			}
		}
	}
	return result
}

// collectAllowed returns tools explicitly allowed (for SDK AllowedTools).
func (p *agentParsedPermissions) collectAllowed() []string {
	var tools []string
	for _, rule := range p.rules {
		if rule.prefix == "allow" {
			tools = append(tools, rule.toolName)
		}
	}
	return tools
}

// collectDenied returns tools explicitly denied (for SDK DisallowedTools/ExcludedTools).
func (p *agentParsedPermissions) collectDenied() []string {
	var tools []string
	for _, rule := range p.rules {
		if rule.prefix == "deny" {
			tools = append(tools, rule.toolName)
		}
	}
	return tools
}

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
	case "codex":
		return newCodexProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", cfg.Agent)
	}
}

func (rnr *agentRunner) Run(ctx context.Context, s *step) error {
	if !scope.IsRunAgentAllowed() {
		return errors.New("scope error: agent runner is not allowed. 'run:agent' scope is required")
	}

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
	if resp == nil {
		return fmt.Errorf("agent provider returned nil response")
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
