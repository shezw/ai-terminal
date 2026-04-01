package llm

import (
	"context"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client is the unified interface for LLM calls.
type Client interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}
