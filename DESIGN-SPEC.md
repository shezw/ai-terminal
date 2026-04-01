# AI-Terminal Design Specification

## 1. Repository Goal
- **Purpose**: 自然语言驱动的命令行终端工具，支持本地模型和远程 API
- **Primary users**: 开发者、系统管理员
- **Repository type**: CLI 应用 + 本地模型服务

## 2. Scope

### In scope (v1.0.0)
- `show` 模式：自然语言 → 命令行建议
- `exec` 模式：自然语言 → 命令执行（含安全目录白/黑名单）
- `model` 子命令：本地模型管理（通过 llama-server 子进程）
- `rem` 子命令：K-V 偏好记忆
- OpenAI 兼容 API 接入
- `@` `@!` `@$` `@!$` 快捷命令（符号链接）
- `<think>` 标签灰色渲染
- `--think` / `--kill` 参数
- 路径统一：所有路径转绝对路径

### Out of scope (v1)
- GUI / TUI 界面
- 多轮对话上下文
- 插件系统

### Key constraints
- 跨平台（macOS / Linux / Windows）
- 纯 Go（llama-server 作为外部子进程管理）

## 3. Architecture

```
┌────────────────────────────────────┐
│           CLI Entry (main)         │
│  ai-terminal $mode [$opts] 'text'  │
├────────┬───────┬────────┬──────────┤
│  show  │ exec  │ model  │   rem    │
├────────┴───────┴────────┴──────────┤
│          LLM Client Layer          │
│   ┌─────────────┬───────────────┐  │
│   │ Local Server │  Remote API  │  │
│   │ (llama-svr)  │ (OpenAI compat)│ │
│   └─────────────┴───────────────┘  │
├────────────────────────────────────┤
│  Config / Path / Safety / Render   │
└────────────────────────────────────┘
```

### Main modules
- **cmd/**: CLI 入口与参数解析
- **mode/**: show / exec / model / rem 四个模式的实现
- **llm/**: LLM 客户端（本地 + 远程统一接口）
- **server/**: llama-server 生命周期管理（下载/启动/停止/配置）
- **config/**: 配置文件读写、路径管理
- **safety/**: exec 模式的白名单/黑名单检查
- **render/**: 终端输出渲染（think 灰色、命令高亮）

### External dependencies
- llama.cpp 预编译 `llama-server` 二进制（按平台下载）
- OpenAI 兼容 API endpoint

## 4. Project Structure

```
ai-terminal/
├── cmd/
│   └── ai-terminal/
│       └── main.go            # 主入口
├── internal/
│   ├── mode/
│   │   ├── show.go            # show 模式
│   │   ├── exec.go            # exec 模式
│   │   ├── model.go           # model 管理
│   │   └── rem.go             # rem K-V 记忆
│   ├── llm/
│   │   ├── client.go          # 统一 LLM 接口
│   │   ├── local.go           # 本地 llama-server 调用
│   │   └── remote.go          # OpenAI 兼容 API 调用
│   ├── server/
│   │   ├── manager.go         # llama-server 生命周期
│   │   ├── download.go        # 模型 & server 二进制下载
│   │   └── template.go        # 模型模板管理
│   ├── config/
│   │   ├── config.go          # 配置读写
│   │   └── paths.go           # 路径规范化
│   ├── safety/
│   │   └── check.go           # 白名单/黑名单校验
│   └── render/
│       └── output.go          # 终端输出（think 灰色等）
├── scripts/
│   ├── install.sh             # 安装脚本（含符号链接）
│   └── wrappers/
│       ├── @.sh               # @ 快捷命令
│       ├── @!.sh              # @! 快捷命令
│       ├── @$.sh              # @$ 快捷命令
│       └── @!$.sh             # @!$ 快捷命令
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── VERSION
├── DESIGN-SPEC.md
└── init.md
```

## 5. Standards
- **Language**: Go 1.22+
- **Formatting**: `gofmt` / `goimports`
- **Test strategy**: `go test ./...`（单元测试优先）
- **Version source**: `VERSION` 文件 (1.0.0)
- **Naming**: snake_case 文件名, CamelCase Go 标识符

## 6. Delivery Model
- **Build**: `make build` → 输出到 `build/`
- **Install**: `make install` → 复制二进制 + 创建符号链接
- **Release**: GitHub Releases（跨平台二进制）

## 7. Collaboration
- **Visibility**: Public
- **Owner**: shezw
- **Remote**: https://github.com/shezw/ai-terminal

## 8. Config File Format

配置目录: `~/.ai-terminal/`

```
~/.ai-terminal/
├── config.yaml          # 主配置（mode, api endpoint, model 等）
├── local-rem            # K-V 偏好存储（简单 key=value 纯文本，便于 grep）
├── models/              # 下载的模型文件
└── templates/           # 模型模板
```

选择 YAML 作为主配置格式（可读性好，LLM 生成/解析友好），`local-rem` 用纯文本 K-V 便于快速读取。

## 9. Exec 安全策略

- **默认白名单**: `~`（用户主目录及子目录）
- `--allow <path>`: 添加路径到白名单
- `--denied <path>`: 添加路径到黑名单
- 黑名单优先于白名单
- 非白名单路径的命令：终端提示并跳过
