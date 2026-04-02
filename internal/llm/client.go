package llm

import (
	"context"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamCallback is called for each streaming chunk.
type StreamCallback func(chunk string)

// Client is the unified interface for LLM calls.
type Client interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message, cb StreamCallback) (string, error)
}
