package mode

import (
	"context"
	"fmt"

	"github.com/shezw/ai-terminal/internal/llm"
	"github.com/shezw/ai-terminal/internal/render"
)

const systemPromptEN = `You are a service assistant designed solely for command-line operations. Please analyze the following content and provide recommended command-line operations with their explanations. Each command should be on its own line.`

const systemPromptZH = `你是一个仅用于命令行操作的服务助手。请根据以下内容做出分析，给出建议的命令行操作和相应的说明。每个指令单独一行。`

func SystemPrompt(lang string) string {
	if lang == "zh" {
		return systemPromptZH
	}
	return systemPromptEN
}

func RunShow(ctx context.Context, client llm.Client, request string, lang string, remKV map[string]string) error {
	messages := buildMessages(lang, request, remKV)

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	output := render.RenderResponse(resp)
	fmt.Print(output)
	return nil
}

func buildMessages(lang string, request string, remKV map[string]string) []llm.Message {
	messages := []llm.Message{
		{Role: "system", Content: SystemPrompt(lang)},
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
