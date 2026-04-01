package render

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	Reset  = "\033[0m"
	Gray   = "\033[90m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

var thinkRegex = regexp.MustCompile(`(?s)<think>(.*?)</think>`)

// RenderResponse processes model output, rendering <think> blocks in gray.
func RenderResponse(text string) string {
	var result strings.Builder
	lastIndex := 0

	matches := thinkRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if match[0] > lastIndex {
			result.WriteString(text[lastIndex:match[0]])
		}
		thinkContent := strings.TrimSpace(text[match[2]:match[3]])
		if thinkContent != "" {
			result.WriteString(Gray)
			result.WriteString("[thinking] ")
			result.WriteString(thinkContent)
			result.WriteString(Reset)
			result.WriteString("\n")
		}
		lastIndex = match[1]
	}

	if lastIndex < len(text) {
		result.WriteString(text[lastIndex:])
	}

	return result.String()
}

func PrintCommand(cmd string) {
	fmt.Printf("%s%s$%s %s\n", Bold, Green, Reset, cmd)
}

func PrintWarning(msg string) {
	fmt.Printf("%s⚠ %s%s\n", Yellow, msg, Reset)
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s✗ %s%s\n", Red, msg, Reset)
}

func PrintInfo(msg string) {
	fmt.Printf("%s%s%s\n", Cyan, msg, Reset)
}
