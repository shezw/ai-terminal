# AI-terminal

这是一个用于命令行提示或执行的 自然语言 终端工具。
他采用本地或远程模型来进行辅助提示或直接执行.

## 技术选型

go语言跨平台

## 功能设计

### 特殊定义或设计

提供1个标准工具程序，1个服务程序，6个二进制程序

- 标准二进制工具程序 `ai-terminal`
接受的标准命令行输入为 `ai-terminal $mode [$options] 'natural request'`

- 专用的标准服务程序 `ai-terminal-server`
本质上是一个本地的模型服务，使用llama.cpp来运行。
他主要管理以下几个内容：
1. 本地模型的下载安装卸载
2. 本地模型的标准模版管理
3. 本地模型的端口号、上下文大小等管理

以下6个包装器用于便捷使用，实质是对 ai-terminal 的 bash 套壳

- 快捷二进制 `@`  等价于 `ai-terminal ...`
- 快捷二进制 `@!` 等价于 `ai-terminal exec ...`
- 快捷二进制 `@$` 等价于 `ai-terminal show --think ...`
- 快捷二进制 `@!$` 等价于 `ai-terminal exec --think ...`

> 缺省功能为 `show`

- 对于 <think> 的处理
当模型恢复中包含 `think` 问题的时候，使用灰色字显示在命令行中.

- 路径统一标准
在程序中的任何部分，涉及到路径一律使用绝对路径，如果用户输入了 相对路径，则根据当前路径和相对关系做换算再继续

### 程序基础功能设计

1. show 提供命令提示作用 (show 模式)，可输出分析、和最终的指令提示。

该功能提供对用户问题的分析和输出，核心提示词为：

中文:
`你是一个仅用于命令行操作的服务助手。请根据以下内容做出分析，给出建议的命令行操作和相应的说明。每个指令单独一行。`

英文:
`You are a service assistant designed solely for command-line operations. Please analyze the following content and provide recommended command-line operations with their explanations. Each command should be on its own line.`

2. exec 提供命令行执行基础文件功能 (exec 模式)，可以配置黑名单、白名单功能，检测到非安全目录拒绝执行，只把命令输出到终端显示

该功能提供对用户问题的分析和输出，并允许执行在安全目录下的文件相关的操作

当 `exec` 模式使用 `--allow xxx` 表示 将 `xxx`作为顶级目录加入白名单
当 `exec` 模式使用 `--denied xxx` 表示 将 `xxx` 作为顶级目录加入黑名单

多条指令使用以下指定格式输出:

```
<icmd>xxx</icmd>
<icmd>xxx</icmd>
```

内容输出时要检测是否有多条指令，如果有，把每一条指令加入到缓冲区。并依次要求用户输入回车来继续。任何一条指令，需要将其中的文件路径读取出来，并与白名单、黑名单对比，若不符合安全定义，则弹出提示并跳过。

3. model 提供 模型服务模式的选择（API/自动安装），模型的管理，包括本地的环境检测或安装。

所有设定项目如下：

- install model 是否安装模型? 是/否
    - 是
        - 检测当前系统性能 并安装合适的模型 （要求根据芯片、系统、显卡状态分3个档次 0.8B 4B/8B 30B）
    - 否
        ... 配置API服务的必要参数

5. rem 提供自定义参数记忆的能力 当前只支持 K-V 

```bash
@ rem 我的网址 https://shezw.com
ai-terminal rem 我的网址 https://shezw.com
```

将偏好存储在 `~/.ai-terminal/local-rem` 中，并且每次都发送给模型

### 参数

1. 启用思考 `--think, -t`

其他 暂缺，保留 `-*, --**` 的格式检测，注意 `@,@@,@!`等命令不支持参数，即所有输入都默认为文本指令

2. 强制关闭 `--kill, -q`

强制关闭本地服务（如果使用本地模型的话），它可用于在模型输出未知异常难移停下来的时候使用


### 其他

空仓库地址: https://github.com/shezw/ai-terminal