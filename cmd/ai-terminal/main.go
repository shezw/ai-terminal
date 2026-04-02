package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/shezw/ai-terminal/internal/config"
	"github.com/shezw/ai-terminal/internal/llm"
	"github.com/shezw/ai-terminal/internal/mode"
	"github.com/shezw/ai-terminal/internal/render"
	"github.com/shezw/ai-terminal/internal/server"
)

var version = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	args := os.Args[1:]
	modeStr, opts, request := parseArgs(args)

	switch modeStr {
	case "show":
		if err := runShowMode(ctx, request, opts); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	case "exec":
		if err := runExecMode(ctx, request, opts); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	case "model":
		if err := mode.RunModel(getRemaining(args)); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	case "rem":
		if err := mode.RunRem(getRemaining(args)); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	case "--kill", "-q":
		if err := server.Stop(); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	case "--version", "-v":
		fmt.Println("ai-terminal", version)
	case "--help", "-h":
		printUsage()
	default:
		request = strings.Join(args, " ")
		if err := runShowMode(ctx, request, &Options{}); err != nil {
			render.PrintError(err.Error())
			os.Exit(1)
		}
	}
}

type Options struct {
	Think  bool
	Allow  []string
	Denied []string
}

func parseArgs(args []string) (string, *Options, string) {
	opts := &Options{}
	modeStr := ""
	var requestParts []string

	i := 0
	if i < len(args) {
		switch args[i] {
		case "show", "exec", "model", "rem", "--kill", "-q", "--version", "-v", "--help", "-h":
			modeStr = args[i]
			i++
		}
	}

	for i < len(args) {
		switch args[i] {
		case "--think", "-t":
			opts.Think = true
		case "--allow":
			if i+1 < len(args) {
				i++
				absPath, err := config.AbsolutePath(args[i])
				if err == nil {
					opts.Allow = append(opts.Allow, absPath)
				}
			}
		case "--denied":
			if i+1 < len(args) {
				i++
				absPath, err := config.AbsolutePath(args[i])
				if err == nil {
					opts.Denied = append(opts.Denied, absPath)
				}
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				requestParts = append(requestParts, args[i])
			}
		}
		i++
	}

	if modeStr == "" {
		modeStr = "show"
	}

	return modeStr, opts, strings.Join(requestParts, " ")
}

func getRemaining(args []string) []string {
	if len(args) > 1 {
		return args[1:]
	}
	return nil
}

func getClient(cfg *config.Config) llm.Client {
	if cfg.Mode == "local" {
		return llm.NewLocalClient(&cfg.Server)
	}
	return llm.NewRemoteClient(&cfg.API)
}

func runShowMode(ctx context.Context, request string, opts *Options) error {
	if request == "" {
		return fmt.Errorf("no request provided")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.Mode == "local" && !server.IsRunning() {
		if err := server.Start(&cfg.Server); err != nil {
			return err
		}
	}

	client := getClient(cfg)

	remKV, err := mode.LoadRem()
	if err != nil {
		remKV = make(map[string]string)
	}

	return mode.RunShow(ctx, client, request, cfg.Language, remKV, opts.Think)
}

func runExecMode(ctx context.Context, request string, opts *Options) error {
	if request == "" {
		return fmt.Errorf("no request provided")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	safetyCfg := cfg.Safety
	safetyCfg.AllowList = append(safetyCfg.AllowList, opts.Allow...)
	safetyCfg.DeniedList = append(safetyCfg.DeniedList, opts.Denied...)

	if cfg.Mode == "local" && !server.IsRunning() {
		if err := server.Start(&cfg.Server); err != nil {
			return err
		}
	}

	client := getClient(cfg)

	remKV, err := mode.LoadRem()
	if err != nil {
		remKV = make(map[string]string)
	}

	return mode.RunExec(ctx, client, request, cfg.Language, &safetyCfg, remKV, opts.Think)
}

func printUsage() {
	fmt.Println("ai-terminal - Natural language terminal assistant")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ai-terminal [mode] [options] 'request'")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  show      Analyze and suggest commands (default)")
	fmt.Println("  exec      Analyze and execute commands with safety checks")
	fmt.Println("  model     Manage local LLM models (install/list/remove)")
	fmt.Println("  rem       Save/list key-value preferences")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --think, -t     Enable thinking output")
	fmt.Println("  --allow <path>  Add path to exec whitelist")
	fmt.Println("  --denied <path> Add path to exec blacklist")
	fmt.Println("  --kill, -q      Force stop local llama-server")
	fmt.Println("  --version, -v   Show version")
	fmt.Println("  --help, -h      Show this help")
	fmt.Println()
	fmt.Println("Shortcuts:")
	fmt.Println("  @               Alias for ai-terminal")
	fmt.Println("  @!              Alias for ai-terminal exec")
	fmt.Println("  @#              Alias for ai-terminal show --think")
	fmt.Println("  @!#             Alias for ai-terminal exec --think")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ai-terminal 'list all docker containers'")
	fmt.Println("  ai-terminal exec 'create a hello.txt file'")
	fmt.Println("  ai-terminal rem my-url https://shezw.com")
	fmt.Println("  ai-terminal model install")
}
