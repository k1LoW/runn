package runn

import "context"

// agentProvider abstracts SDK-specific agent implementations.
type agentProvider interface {
	// Run sends a prompt and returns the agent's text response.
	// Conversation context is managed by the SDK session internally.
	// When clearContext is true, the provider resets the session.
	Run(ctx context.Context, req *agentRunRequest) (*AgentResponse, error)
	Close() error
}

type agentRunRequest struct {
	prompt       string
	clearContext bool
}

// AgentResponse is the response from an agent provider.
type AgentResponse struct {
	Content string
}
