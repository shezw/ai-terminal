package render

import (
	"strings"
	"testing"
)

func TestRenderResponse_NoThinkTags(t *testing.T) {
	input := "Just a normal response"
	result := RenderResponse(input)
	if result != input {
		t.Errorf("expected unchanged output, got %q", result)
	}
}

func TestRenderResponse_WithThinkBlock(t *testing.T) {
	input := "<think>reasoning here</think>The answer is 42."
	result := RenderResponse(input)
	if !strings.Contains(result, "[thinking]") {
		t.Error("expected [thinking] label in output")
	}
	if !strings.Contains(result, "reasoning here") {
		t.Error("expected think content in output")
	}
	if !strings.Contains(result, "The answer is 42.") {
		t.Error("expected answer in output")
	}
}

func TestRenderResponse_EmptyThinkBlock(t *testing.T) {
	input := "<think></think>Answer"
	result := RenderResponse(input)
	if strings.Contains(result, "[thinking]") {
		t.Error("empty think block should not produce [thinking] label")
	}
	if !strings.Contains(result, "Answer") {
		t.Error("expected answer in output")
	}
}

func TestStreamState_PlainText(t *testing.T) {
	state := NewStreamState()
	state.RenderChunk("hello world")
	if state.InThink {
		t.Error("should not be in think mode")
	}
}

func TestStreamState_ThinkTransition(t *testing.T) {
	state := NewStreamState()
	state.RenderChunk("<think>")
	if !state.InThink {
		t.Error("should be in think mode after <think> tag")
	}
	state.RenderChunk("reasoning")
	state.RenderChunk("</think>")
	if state.InThink {
		t.Error("should exit think mode after </think> tag")
	}
}
