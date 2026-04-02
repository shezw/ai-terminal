package mode

import (
	"context"
	"fmt"

	"github.com/shezw/ai-terminal/internal/llm"
	"github.com/shezw/ai-terminal/internal/render"
)

const systemPromptEN = "You are a service assistant designed solely for command-line operations. Please analyze the following content and provide recommended command-line operations with their explanations. Each command should be on its own line."

const systemPromptZH = "\u4f60\u662f\u4e00\u4e2a\u4ec5\u7528\u4e8e\u547d\u4ee4\u884c\u64cd\u4f5c\u7684\u670d\u52a1\u52a9\u624b\u3002\u8bf7\u6839\u636e\u4ee5\u4e0b\u5185\u5bb9\u505a\u51fa\u5206\u6790\uff0c\u7ed9\u51fa\u5efa\u8bae\u7684\u547d\u4ee4\u884c\u64cd\u4f5c\u548c\u76f8\u5e94\u7684\u8bf4\u660e\u3002\u6bcf\u4e2a\u6307\u4ee4\u5355\u72ec\u4e00\u884c\u3002"

func SystemPrompt(lang string) string {
	if lang == "zh" {
		return systemPromptZH
	}
	return systemPromptEN
}

func RunShow(ctx context.Context, client llm.Client, request string, lang string, remKV map[string]string, think bool) error {
	messages := buildMessages(lang, request, remKV, think)

	// Use streaming
	state := render.NewStreamState()
	_, err := client.ChatStream(ctx, messages, func(chunk string) {
		state.RenderChunk(chunk)
	})
	state.Flush()

	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	return nil
}

func buildMessages(lang string, request string, remKV map[string]string, think bool) []llm.Message {
	sysPrompt := SystemPrompt(lang)
	if think {
		sysPrompt += "\nPlease think step by step before answering."
	}

	messages := []llm.Message{
		{Role: "system", Content: sysPrompt},
	}

	if len(remKV) > 0 {
		var ctxStr string
		for k, v := range remKV {
			ctxStr += fmt.Sprintf("- %s: %s\n", k, v)
		}
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "User preferences:\n" + ctxStr,
		})
	}

	messages = append(messages, llm.Message{
		Role:    "user",
		Content: request,
	})

	return messages
}
