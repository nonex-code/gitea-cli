# gitea-cli

一个用于操作自部署 Gitea 服务器的命令行工具，使用 Go 编写。**专为 AI agent 设计**。

> **🎉 本项目完全由 AI 自主完成**：包括需求分析、架构设计、代码开发、单元测试、端到端测试（针对真实 Gitea 服务器）以及文档编写，全程由 AI（Claude）独立完成，无人工编写代码。

[English](README.md) | 中文（当前）

## 特性

- 基于官方 [Gitea Go SDK](https://code.gitea.io/sdk/gitea)
- **凭据通过 git 自身的密钥箱（credential helper）管理**，不自行存储密码
- 支持仓库、Issue、PR、用户、组织、团队、Release、Webhook 等资源的完整增删查改
- 配置与凭据分离：服务器信息存配置文件，令牌存密钥箱
- **AI 友好**：全局 `--json` 标志输出结构化 JSON（含错误），非交互环境下自动拒绝需确认的删除操作

## 安装

### 从源码构建

```bash
go build -o gitea-cli .
# 可选：安装到 PATH
go install
```

### 从 GitHub Release 下载

本仓库通过 GitHub Actions 自动构建跨平台二进制。发布新版本时，会生成 Linux / macOS / Windows 等多平台的可执行文件并上传到 GitHub Release。

```bash
# 下载对应平台的二进制（以 v1.0.0、linux/amd64 为例）
gh release download v1.0.0 --pattern 'gitea-cli_v1.0.0_linux_amd64.tar.gz'
tar xzf gitea-cli_v1.0.0_linux_amd64.tar.gz
sudo mv gitea-cli /usr/local/bin/
```

## 认证

gitea-cli 访问 Gitea API 需要认证。**认证的本质很简单：由外部提供访问令牌（token），工具将其通过 HTTP Basic Auth 发送给 Gitea 服务器。**

所谓「初始化」，就是**提供 token 的其中一种方式**——把 token 存入 git 密钥箱，之后每次调用自动读取，无需重复输入。初始化、配置、认证，三者的最终目的都是同一个：让 gitea-cli 能通过 Gitea 服务器的身份验证。

### 认证机制

工具通过 **HTTP Basic Auth** 认证：用户名作为用户名，访问令牌（token）作为密码发送。

采用 Basic Auth 而非 `Authorization: token` 头的原因是为了兼容性——部分部署环境（如前置 WAF / 雷池 SafeLine）会拦截 `Authorization: token` 头，但放行 Basic Auth。

### 令牌的来源（按优先级）

工具按以下优先级查找 token，找到即用，全部未提供则认证失败：

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | 配置文件 `token` 字段 / `GITEA_TOKEN` 环境变量 | 显式写在配置或环境变量中 |
| 2 | git 密钥箱（credential helper） | 通过 `git credential fill` 获取密码作为 token |

### 初始化（把 token 写入密钥箱）

`config init` 把 token 存入 git 密钥箱。这是推荐方式——token 不落盘到配置文件，安全性更好。

```bash
# 交互式：按提示输入服务器地址、用户名和 token
gitea-cli config init

# 非交互式（推荐 AI agent 使用）：token 从 stdin 传入，避免出现在 shell 历史
echo "$TOKEN" | gitea-cli config init --url https://gitea.example.com --username alice --token-stdin

# 或通过环境变量提供服务器地址和用户名
export GITEA_URL=https://gitea.example.com
export GITEA_USERNAME=alice
printf '%s' "$TOKEN" | gitea-cli config init --token-stdin
```

token 通过 `git credential approve` 存入 git 密钥箱（如 macOS 钥匙串、Linux Secret Service、或 `~/.git-credentials`），具体后端由 `git config credential.helper` 决定。**token 不会写入配置文件**。

### 密钥箱管理

```bash
# 查看当前配置与认证状态
gitea-cli config show

# 清除密钥箱中的凭据
gitea-cli config clear
```

## 配置

配置与认证分离：**服务器地址和用户名存在配置文件里，token 存在密钥箱里**（也可选写进配置文件，但不推荐）。

### 配置文件

保存在 `~/.gitea-cli/config.yaml`：

```yaml
url: https://gitea.example.com
username: yourname
# token: （可选，一般不写，改用密钥箱）
```

### 环境变量

优先级高于配置文件：

| 变量 | 说明 |
|------|------|
| `GITEA_URL` | 服务器地址 |
| `GITEA_USERNAME` | 用户名 |
| `GITEA_TOKEN` | 访问令牌（可选） |

## 使用

### 仓库

```bash
gitea-cli repo list                    # 列出我的仓库
gitea-cli repo list username           # 列出某用户/组织的仓库
gitea-cli repo info owner/repo         # 查看详情
gitea-cli repo create my-repo --description "描述" --private
gitea-cli repo delete owner/repo       # 需确认，可 --yes 跳过
gitea-cli repo star owner/repo         # 星标
gitea-cli repo unstar owner/repo       # 取消星标
gitea-cli repo watch owner/repo        # 关注
gitea-cli repo unwatch owner/repo      # 取消关注

# 标签
gitea-cli repo label list owner/repo
gitea-cli repo label create owner/repo "bug" "#ff0000" --description "缺陷"
```

### Issue

```bash
gitea-cli issue list owner/repo --state open
gitea-cli issue info owner/repo 1
gitea-cli issue create owner/repo "标题" --body "内容"
gitea-cli issue close owner/repo 1
gitea-cli issue reopen owner/repo 1
gitea-cli issue delete owner/repo 1    # 需确认，可 --yes 跳过

# 评论
gitea-cli issue comment list owner/repo 1
gitea-cli issue comment add owner/repo 1 "评论内容"
```

### Pull Request

```bash
gitea-cli pr list owner/repo
gitea-cli pr info owner/repo 3
gitea-cli pr create owner/repo "标题" --head feature-branch --base main
```

### 用户

```bash
gitea-cli user info              # 查看当前用户
gitea-cli user info username     # 查看指定用户
gitea-cli user search 关键字
```

### 组织与团队

```bash
# 组织
gitea-cli org list              # 列出我的组织
gitea-cli org list username     # 列出某用户的组织
gitea-cli org info orgname
gitea-cli org create neworg --description "描述"
gitea-cli org delete neworg      # 需确认，可 --yes 跳过

# 组织成员
gitea-cli org member list orgname
gitea-cli org member add orgname username --team Owners
gitea-cli org member remove orgname username

# 团队
gitea-cli org team list orgname
gitea-cli org team create orgname devteam --permission write
gitea-cli org team member list <team-id>
gitea-cli org team member add <team-id> username
gitea-cli org team member remove <team-id> username
```

### Release

```bash
gitea-cli release list owner/repo
gitea-cli release create owner/repo v1.0.0 --title "v1.0.0" --note "发布说明"
gitea-cli release delete owner/repo v1.0.0   # 需确认，可 --yes 跳过
```

### Webhook

```bash
gitea-cli webhook list owner/repo
gitea-cli webhook create owner/repo gitea https://example.com/hook --secret xxx
gitea-cli webhook delete owner/repo <hook-id>   # 需确认，可 --yes 跳过
```

## 面向 AI agent 的使用指南

本工具针对 AI agent（如 Claude Code、pi、Cursor 等）做了专门优化。

### 核心原则

1. **始终携带 `--json` 标志**，所有命令输出结构化 JSON，便于程序化解析。
2. **删除类操作需 `--yes` 标志**：在非交互环境（AI agent）下，没有 `--yes` 会直接报错退出，而不是挂起等待输入。
3. **通过退出码判断成败**：`0` 表示成功，非 `0` 表示失败。

### 输出格式约定

- **成功**：输出 JSON 数据（对象或数组），退出码 `0`。
- **失败**：输出 `{"ok": false, "error": "..."}` 到 stderr，退出码 `1`。

### 非交互初始化

AI agent 无交互终端，需通过参数或环境变量提供信息（详见上文 [认证](#认证) 章节）：

```bash
echo "$TOKEN" | gitea-cli config init --url https://gitea.example.com --username alice --token-stdin
```

### 推荐用法示例

```bash
# 列出仓库（JSON）
gitea-cli repo list --json

# 查看仓库详情
gitea-cli repo info owner/repo --json

# 创建 Issue
gitea-cli issue create owner/repo "标题" --body "内容" --json

# 非交互删除（必须 --yes）
gitea-cli repo delete owner/repo --yes
```

## 目录结构

```
.
├── main.go                 # 入口
├── cmd/                    # CLI 命令定义
│   ├── root.go             # 根命令
│   ├── config.go           # config 子命令
│   ├── repo.go             # 仓库管理（含标签、星标、关注）
│   ├── issue.go            # Issue 管理（含评论）
│   ├── pr.go               # PR 管理
│   ├── user.go             # 用户管理
│   ├── org.go              # 组织管理（含成员）
│   ├── org_team.go         # 组织团队管理
│   ├── release.go          # Release 管理
│   ├── webhook.go          # Webhook 管理
│   ├── helpers.go          # 辅助函数
│   ├── confirm.go          # 删除确认
│   └── password.go         # 终端密码读取
├── config/                 # 配置加载/保存
├── client/                 # Gitea 客户端封装
└── credential/             # git credential 密钥箱交互
```

## 发布

本项目通过 GitHub Actions + GoReleaser 自动化发布。工作流在推送 `v*` 标签时触发，自动完成：

1. 运行单元测试
2. 交叉编译多平台二进制（Linux / macOS / Windows / FreeBSD / OpenBSD × amd64 / arm64 / arm）
3. 创建 GitHub Release 并上传构建好的二进制、校验和与源码包

### 发布一个新版本（使用 gh）

```bash
# 1. 创建并推送版本标签（触发工作流）
git tag v1.0.0
git push origin v1.0.0

# 2. 等待 GitHub Actions 构建完成（约 1-2 分钟），可查看进度：
gh run watch

# 3. 构建完成后，Release 已自动创建并附带二进制；查看：
gh release view v1.0.0

# 4.（可选）列出已上传的资产
gh release view v1.0.0 --json assets
```

> 提示：工作流使用 GoReleaser，版本号取自 git 标签。二进制会注入版本、提交哈希和构建日期，可通过 `gitea-cli --version` 查看。

## 依赖

本项目基于以下优秀开源项目构建：

| 依赖 | 用途 | 链接 |
|------|------|------|
| Gitea Go SDK | Gitea API 客户端 | https://code.gitea.io/sdk/gitea |
| Cobra | CLI 框架 | https://github.com/spf13/cobra |
| Viper | 配置管理 | https://github.com/spf13/viper |
| golang.org/x/term | 终端检测 | https://pkg.go.dev/golang.org/x/term |

## 致谢

感谢以下项目与社区：

- **[Gitea](https://gitea.io/)** 官方团队 —— 提供优秀的自托管 Git 服务，以及官方维护的 [Gitea Go SDK](https://code.gitea.io/sdk/gitea)，本项目正是基于该 SDK 构建。
- **[Cobra](https://github.com/spf13/cobra)** —— 强大的 Go CLI 框架，让命令行工具的构建变得简洁优雅。
- **[Viper](https://github.com/spf13/viper)** —— 灵活的 Go 配置管理库。
- **git credential helper 生态** —— 本项目复用 git 自身的密钥箱机制（libsecret / osxkeychain / pass / store 等）管理凭据，遵循 git 的安全规范。

## 许可证

MIT
