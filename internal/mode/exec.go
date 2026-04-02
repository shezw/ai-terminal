package mode

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/llm"
	"github.com/shezw/ai-terminal/internal/render"
	"github.com/shezw/ai-terminal/internal/safety"
)

const execSystemPromptEN = "You are a service assistant designed solely for command-line operations. Analyze the following request and provide the exact commands to execute. Wrap each command in <icmd>command</icmd> tags. Each command on its own line."

const execSystemPromptZH = "\u4f60\u662f\u4e00\u4e2a\u4ec5\u7528\u4e8e\u547d\u4ee4\u884c\u64cd\u4f5c\u7684\u670d\u52a1\u52a9\u624b\u3002\u5206\u6790\u4ee5\u4e0b\u8bf7\u6c42\u5e76\u63d0\u4f9b\u9700\u8981\u6267\u884c\u7684\u7cbe\u786e\u547d\u4ee4\u3002\u6bcf\u6761\u547d\u4ee4\u7528 <icmd>command</icmd> \u6807\u7b7e\u5305\u88f9\uff0c\u6bcf\u6761\u547d\u4ee4\u72ec\u7acb\u4e00\u884c\u3002"

var icmdRegex = regexp.MustCompile(`<icmd>(.*?)</icmd>`)
var pathRegex = regexp.MustCompile(`(?:^|[\s"'])(/[^\s"']+|~[^\s"']*)`)

func ExecSystemPrompt(lang string) string {
	if lang == "zh" {
		return execSystemPromptZH
	}
	return execSystemPromptEN
}

func RunExec(ctx context.Context, client llm.Client, request string, lang string, cfg *config.ExecSafety, remKV map[string]string, think bool) error {
	sysPrompt := ExecSystemPrompt(lang)
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

	// Streaming: collect full response while displaying think blocks
	state := render.NewStreamState()
	resp, err := client.ChatStream(ctx, messages, func(chunk string) {
		state.RenderChunk(chunk)
	})
	state.Flush()

	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	// Extract commands from full response
	commands := icmdRegex.FindAllStringSubmatch(resp, -1)
	if len(commands) == 0 {
		return nil
	}

	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)

	for i, match := range commands {
		cmd := strings.TrimSpace(match[1])
		if cmd == "" {
			continue
		}

		if !checkCommandSafety(cmd, cfg) {
			render.PrintWarning(fmt.Sprintf("Skipped (unsafe path): %s", cmd))
			continue
		}

		render.PrintCommand(cmd)

		if i < len(commands)-1 {
			fmt.Printf("Press Enter to execute (or 'q' to cancel remaining)... ")
		} else {
			fmt.Printf("Press Enter to execute (or 'q' to cancel)... ")
		}

		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "q" || input == "Q" {
			render.PrintInfo("Cancelled remaining commands.")
			break
		}

		execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		if err := execCmd.Run(); err != nil {
			render.PrintError(fmt.Sprintf("Command failed: %s", err))
		}
		fmt.Println()
	}

	return nil
}

func checkCommandSafety(cmd string, cfg *config.ExecSafety) bool {
	matches := pathRegex.FindAllStringSubmatch(cmd, -1)
	for _, match := range matches {
		p := strings.TrimSpace(match[1])
		if strings.HasPrefix(p, "~") {
			p = strings.Replace(p, "~", config.HomeDir(), 1)
		}
		absPath, err := config.AbsolutePath(p)
		if err != nil {
			continue
		}
		if !safety.IsPathSafe(absPath, cfg) {
			return false
		}
	}
	return true
}
