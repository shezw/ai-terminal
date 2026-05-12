# 测试用例

- 测试 ID: `01`
- 项目: `ai-terminal`
- 场景: `model-config`
- 目标: `internal/mode`
- 创建时间: `2026-05-10 18:30`

## 测试验证

1. 执行 `go test ./...`
2. 执行 `printf 'api\nhttps://example.com/v1\nsk-test\nqwen-test\n' | go run ./cmd/ai-terminal model config`

## 结果判定

- 返回值: `true | false`
- 通过条件: 单元测试通过，且交互式命令将输入写入配置文件
- 失败处理: 更新 `results.db` 中的 `fails`

## 执行记录

- 最近结果: `true`
- 说明: `go test ./...` 通过；交互式命令在隔离 HOME 下成功写出配置文件。首次冒烟因 Go 缓存切换触发依赖下载超时，固定缓存后复跑通过。

