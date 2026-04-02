# AI Terminal

自然语言驱动的命令行终端工具。支持本地模型（llama.cpp）和远程 API（OpenAI 兼容）。

## Features

- **show** — 自然语言 → 命令行建议（流式输出）
- **exec** — 自然语言 → 安全执行命令（白名单/黑名单）
- **model** — 本地模型管理（自动下载 Qwen3 GGUF）
- **rem** — K-V 偏好记忆
- **--think** — 启用思考链输出

## One-Click Install

```bash
curl -fsSL https://raw.githubusercontent.com/shezw/ai-terminal/main/scripts/quick-install.sh | bash
```

需要预编译的 release 发布到 GitHub Releases。

## Build from Source

```bash
# Build
make build

# Install (copies binary + shortcuts to /usr/local/bin)
sudo make install

# Configure API
mkdir -p ~/.ai-terminal
cat > ~/.ai-terminal/config.yaml << EOF
mode: api
api:
  endpoint: https://api.openai.com/v1
  api_key: sk-your-key
  model: gpt-4o-mini
language: zh
EOF

# Or use a local model
ai-terminal model install
```

## Usage

```bash
ai-terminal '列出所有运行中的docker容器'
ai-terminal exec '创建一个hello.txt文件'
ai-terminal show --think '分析磁盘用量'
ai-terminal exec --think '清理docker缓存'
```

## Shortcuts

| Command | Equivalent |
|---------|-----------|
| `@` | `ai-terminal` |
| `@!` | `ai-terminal exec` |
| `@#` | `ai-terminal show --think` |
| `@!#` | `ai-terminal exec --think` |

## Safety

`exec` 模式默认白名单为 `~`（用户主目录）。

```bash
ai-terminal exec --allow /tmp '...'       # 添加白名单
ai-terminal exec --denied /etc '...'      # 添加黑名单
```

## Memory

```bash
ai-terminal rem 我的网址 https://shezw.com
ai-terminal rem                            # 列出所有
```

## License

MIT
