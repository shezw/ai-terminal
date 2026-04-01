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

const execSystemPromptEN = `You are a service assistant designed solely for command-line operations. Analyze the following request and provide the exact commands to execute. Wrap each command in <icmd>command</icmd> tags. Each command on its own line.`

const execSystemPromptZH = `你是一个仅用于命令行操作的服务助手。分析以下请求并提供需要执行的精确命令。每条命令用 <icmd>command</icmd> 标签包裹，每条命令独立一行。`

var icmdRegex = regexp.MustCompile(`<icmd>(.*?)</icmd>`)
var pathRegex = regexp.MustCompile(`(?:^|[\s"'])(/[^\s"']+|~[^\s"']*)`)

func ExecSystemPrompt(lang string) string {
	if lang == "zh" {
		return execSystemPromptZH
	}
	return execSystemPromptEN
}

func RunExec(ctx context.Context, client llm.Client, request string, lang string, cfg *config.ExecSafety, remKV map[string]string) error {
	messages := []llm.Message{
		{Role: "system", Content: ExecSystemPrompt(lang)},
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

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	output := render.RenderResponse(resp)

	commands := icmdRegex.FindAllStringSubmatch(output, -1)
	if len(commands) == 0 {
		fmt.Print(output)
		return nil
	}

	cleanOutput := icmdRegex.ReplaceAllString(output, "")
	cleanOutput = strings.TrimSpace(cleanOutput)
	if cleanOutput != "" {
		fmt.Println(cleanOutput)
		fmt.Println()
	}

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
